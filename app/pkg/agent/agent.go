package agent

import (
	"connectrpc.com/connect"
	"context"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1/v1alpha1connect"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"io"
	"strings"

	"github.com/jlewi/foyle/app/pkg/llms"

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
	v1alpha1connect.UnimplementedAIServiceHandler
	completer llms.Completer
	config    config.Config
	db        *learn.InMemoryExampleDB
}

func NewAgent(cfg config.Config, completer llms.Completer, inMemoryExampleDB *learn.InMemoryExampleDB) (*Agent, error) {
	if cfg.Agent == nil {
		return nil, errors.New("Configuration is missing AgentConfig; configuration must define the agent field.")
	}
	log := zapr.NewLogger(zap.L())
	log.Info("Creating agent", "config", cfg.Agent)

	if completer == nil {
		return nil, errors.New("Completer is required")
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
		completer: completer,
		config:    cfg,
		db:        inMemoryExampleDB,
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

		blocks, err := a.completer.Complete(ctx, systemPrompt, sb.String())

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

		return blocks, nil
	}
	err := errors.Errorf("Failed to generate a chat completion after %d tries", maxTries)
	log.Error(err, "Failed to generate a chat completion", "maxTries", maxTries)
	return nil, err
}

func (a *Agent) StreamGenerate(ctx context.Context, stream *connect.BidiStream[v1alpha1.StreamGenerateRequest, v1alpha1.StreamGenerateResponse]) error {
	log := logs.FromContext(ctx)
	log.Info("Agent.StreamGenerate")
	notebookUri := ""
	var selectedCell int32
	reqCount := 0

	for {
		req, err := stream.Receive()

		if err != nil {
			if errors.Is(err, io.EOF) {
				// The client has closed the stream
				log.Info("Client closed the stream")
				return nil
			}
			if errors.Is(err, context.Canceled) {
				// The context was cancelled (e.g., client disconnected)
				log.Info("Stream context cancelled")
				return nil
			}
			// Some other error occurred
			log.Error(err, "Error receiving from stream")
			return err
		}
		reqCount++

		isValidErr := func() error {
			if reqCount == 1 {
				if req.GetFullContext() == nil {
					return status.Errorf(codes.InvalidArgument, "First request must have a full context")
				}
				log.Info("Received full context", "context", req.GetFullContext())
				if req.GetFullContext().GetNotebookUri() == "" {
					return status.Errorf(codes.InvalidArgument, "First request must have a notebookUri")
				}
				if req.GetFullContext().GetSelected() < 0 {
					return status.Errorf(codes.InvalidArgument, "First request must have a selected cell")
				}
				notebookUri = req.GetFullContext().GetNotebookUri()
				selectedCell = req.GetFullContext().GetSelected()
			} else {
				if req.GetUpdate() == nil {
					return status.Errorf(codes.InvalidArgument, "Every request except the first one must have an update")
				}
			}
			return nil
		}()

		if isValidErr != nil {
			log.Info("Request is invalid", "err", isValidErr)
			return isValidErr
		}

		b, err := protojson.Marshal(req)
		if err != nil {
			log.Error(err, "Failed to marshal request")
			return err
		}

		// TODO(jeremy): Get rid of this its only for debugging.
		log.Info("Received request", "request", string(b))
		// Process the request and generate a response
		// This is where you'd implement your AI logic
		response := &v1alpha1.StreamGenerateResponse{
			// Fill in the response fields based on your protobuf definition
			Blocks: []*v1alpha1.Block{
				{
					Contents: "Generated text based on: " + string(b),
				},
			},

			NotebookUri: notebookUri,
			InsertAt:    selectedCell + 1,
		}
		log.Info("Sending response", "notebook", response.GetNotebookUri(), "insertAt", response.GetInsertAt(), "numBlocks", len(response.GetBlocks()))
		if err := stream.Send(response); err != nil {
			return err
		}
	}
}

func (a *Agent) Simple(ctx context.Context, req *connect.Request[v1alpha1.StreamGenerateRequest]) (*connect.Response[v1alpha1.StreamGenerateResponse], error) {
	log := logs.FromContext(ctx)
	log.Info("Agent.Simple")

	b, err := protojson.Marshal(req.Msg)
	if err != nil {
		log.Error(err, "Failed to marshal request")
		return nil, err
	}
	// Process the request and generate a response
	// This is where you'd implement your AI logic
	response := &v1alpha1.StreamGenerateResponse{
		// Fill in the response fields based on your protobuf definition
		Blocks: []*v1alpha1.Block{
			{
				Contents: "Generated text based on: " + string(b),
			},
		},
	}
	res := connect.NewResponse(response)
	res.Header().Set("AIService-Version", "v1alpha1")
	return res, nil
}
