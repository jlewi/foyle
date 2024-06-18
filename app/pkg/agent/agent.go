package agent

import (
	"context"
	"strings"

	"github.com/jlewi/foyle/app/pkg/learn"

	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"

	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/docs"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/app/pkg/oai"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
)

const (
	maxTries = 3
	// MaxDocChars is an upper limit for the number of characters to include in prompts to avoid hitting
	// OpenAI's context length limits. This can be an upper bound because if we get a context length exceeded
	// error the code will automatically try to shrink the document even further.
	// We use the heuristic 1 token ~ 2 characters
	// We are currently using GPT3.5 which has a context window of 16385 tokens.
	// (https://platform.openai.com/docs/models/gpt-3-5-turbo)
	// If we use 50% of that's 16000 characters.
	MaxDocChars = 16000
	temperature = 0.9
)

// Agent is the agent.
type Agent struct {
	v1alpha1.UnimplementedGenerateServiceServer
	client *openai.Client
	config config.Config
	db     *learn.InMemoryExampleDB
}

func NewAgent(cfg config.Config, client *openai.Client, inMemoryExampleDB *learn.InMemoryExampleDB) (*Agent, error) {
	if cfg.Agent == nil {
		return nil, errors.New("Configuration is missing AgentConfig; configuration must define the agent field.")
	}
	log := zapr.NewLogger(zap.L())
	log.Info("Creating agent", "config", cfg.Agent)

	if client == nil {
		return nil, errors.New("OpenAI client is required")
	}
	if cfg.Agent.RAG != nil && cfg.Agent.RAG.Enabled {
		if inMemoryExampleDB == nil {
			return nil, errors.New("RAG is enabled but learn is nil; learn must be set to use RAG")
		}
		log.Info("RAG is enabled")
	} else {
		inMemoryExampleDB = nil
	}

	return &Agent{
		client: client,
		config: cfg,
		db:     inMemoryExampleDB,
	}, nil
}

func (a *Agent) Generate(ctx context.Context, req *v1alpha1.GenerateRequest) (*v1alpha1.GenerateResponse, error) {
	span := trace.SpanFromContext(ctx)
	log := logs.FromContext(ctx)
	traceId := span.SpanContext().TraceID()
	log = log.WithValues("traceId", traceId, "evalMode", a.config.EvalMode())
	ctx = logr.NewContext(ctx, log)

	var examples []*v1alpha1.Example
	if a.config.UseRAG() {
		var err error
		examples, err = a.db.GetExamples(ctx, req.Doc, a.config.RagMaxResults())
		if err != nil {
			// Fail gracefully; keep going without examples
			log.Error(err, "Failed to get examples")
			examples = nil
		}
	}

	log.Info("Agent.Generate", zap.Object("request", req))
	blocks, err := a.completeWithRetries(ctx, req, examples)
	if err != nil {
		// TODO(jeremy): Should we set a status code?
		log.Error(err, "Agent.Generate failed to generate completions")
		return nil, err
	}

	// Attach block ids to any blocks generated.
	// N.B. This is kind of a last resort to make sure all blocks have an ID set. In general, we want to set blockIds
	// earlier in the processing pipeline so that any log messages involving blocks has block ids set. BlockIDs
	// should get set in parseResponse. When block Ids are first set.
	blockIds, err := docs.SetBlockIds(blocks)
	if err != nil {
		log.Error(err, "Agent.Generate, failed to set block ids", "blocks", blocks, "blockIds", blockIds)
	} else {
		log.Info("Agent.Generate returning blocks", "blockIds", blockIds)
	}

	resp := &v1alpha1.GenerateResponse{
		Blocks:  blocks,
		TraceId: traceId.String(),
	}

	log.Info("Agent.Generate returning response", zap.Object("response", resp))
	return resp, nil
}

func (a *Agent) completeWithRetries(ctx context.Context, req *v1alpha1.GenerateRequest, examples []*v1alpha1.Example) ([]*v1alpha1.Block, error) {
	log := logs.FromContext(ctx)

	t := docs.NewTailer(req.Doc.GetBlocks(), MaxDocChars)

	exampleArgs := make([]Example, 0, len(examples))
	for _, example := range examples {
		exampleArgs = append(exampleArgs, Example{
			Input:  docs.DocToMarkdown(example.Query),
			Output: docs.BlocksToMarkdown(example.Answer),
		})
	}
	for try := 0; try < maxTries; try++ {
		args := promptArgs{
			Document: t.Text(),
			Examples: exampleArgs,
		}

		var sb strings.Builder
		if err := promptTemplate.Execute(&sb, args); err != nil {
			return nil, errors.Wrapf(err, "Failed to execute prompt template")
		}

		messages := []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{Role: openai.ChatMessageRoleUser,
				Content: sb.String(),
			},
		}
		request := openai.ChatCompletionRequest{
			Model:       a.config.GetModel(),
			Messages:    messages,
			MaxTokens:   2000,
			Temperature: temperature,
		}

		log.Info("OpenAI:CreateChatCompletion", "request", request)
		resp, err := a.client.CreateChatCompletion(ctx, request)

		if err != nil {
			if oai.ErrorIs(err, oai.ContextLengthExceededCode) {
				log.Info("OpenAI:ContextLengthExceeded", "err", err)
				if !t.Shorten() {
					return nil, errors.Wrapf(err, "the document can't be shortened any further to fit within the context window")
				}
				continue
			}
			// TODO(jeremy): Should we surface the error to the user as blocks in the notebook
			return nil, errors.Wrapf(err, "CreateChatCompletion failed")
		}

		log.Info("OpenAI:CreateChatCompletion response", "resp", resp)

		blocks, err := a.parseResponse(ctx, &resp)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to parse response")
		}

		return blocks, nil
	}
	err := errors.Errorf("Failed to generate a chat completion after %d tries", maxTries)
	log.Error(err, "Failed to generate a chat completion", "maxTries", maxTries)
	return nil, err
}

func (a *Agent) parseResponse(ctx context.Context, resp *openai.ChatCompletionResponse) ([]*v1alpha1.Block, error) {
	log := logs.FromContext(ctx)
	allBlocks := make([]*v1alpha1.Block, 0, 10)
	for _, choice := range resp.Choices {
		if choice.Message.Content == "" {
			continue
		}

		blocks, err := docs.MarkdownToBlocks(choice.Message.Content)
		if err != nil {
			log.Error(err, "Failed to parse markdown to blocks", "markdown", choice.Message.Content)
			b := &v1alpha1.Block{
				Kind:     v1alpha1.BlockKind_MARKUP,
				Contents: choice.Message.Content,
			}
			allBlocks = append(allBlocks, b)
			continue
		}

		// Set block ids
		if _, err := docs.SetBlockIds(blocks); err != nil {
			return nil, errors.Wrapf(err, "Failed to set block ids")
		}

		allBlocks = append(allBlocks, blocks...)
	}
	return allBlocks, nil
}
