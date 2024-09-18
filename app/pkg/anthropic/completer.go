package anthropic

import (
	"context"
	"net/http"

	"github.com/jlewi/foyle/app/api"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/docs"
	"github.com/jlewi/foyle/app/pkg/llms"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/liushuangls/go-anthropic/v2"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

const (
	temperature = 0.9
)

func NewCompleter(cfg config.Config, client *anthropic.Client) (*Completer, error) {
	return &Completer{
		client: client,
		config: cfg,
	}, nil
}

// Completer is a wrapper around OpenAI that implements the Completer interface.
type Completer struct {
	client *anthropic.Client
	config config.Config
}

// Complete returns a ContextLengthExceededError if the context is too long
func (c *Completer) Complete(ctx context.Context, systemPrompt string, message string) ([]*v1alpha1.Block, error) {
	tp := tracer()
	// Start a span to record metrics.
	ctx, span := tp.Start(ctx, "Complete", trace.WithAttributes(attribute.String("llm.model", c.config.GetModel()), attribute.String("llm.provider", string(api.ModelProviderAnthropic))))
	defer span.End()

	log := logs.FromContext(ctx)
	// See: https://docs.anthropic.com/en/api/messages
	// Claude doesn't have a system prompt.
	// First message must also be a user message.

	messages := []anthropic.Message{
		{Role: anthropic.RoleUser,
			Content: []anthropic.MessageContent{
				{Type: anthropic.MessagesContentTypeText,
					Text: proto.String(message),
				},
			},
		},
	}

	request := anthropic.MessagesRequest{
		Model:       c.config.GetModel(),
		Messages:    messages,
		MaxTokens:   2000,
		Temperature: proto.Float32(temperature),
		System:      systemPrompt,
	}

	log.Info("Anthropic:CreateMessages", "request", request)
	resp, err := c.client.CreateMessages(ctx, request)

	if err != nil {
		// https://docs.anthropic.com/en/api/errors
		aErr, ok := err.(*anthropic.RequestError)
		if ok {
			span.SetAttributes(attribute.Int("llm.statusCode", aErr.StatusCode))
		}
		// 413 means context length exceeded.
		if ok && aErr.StatusCode == http.StatusRequestEntityTooLarge {
			return nil, llms.ContextLengthExceededError{Cause: err}
		}
		// TODO(jeremy): Should we surface the error to the user as blocks in the notebook
		return nil, errors.Wrapf(err, "CreateChatCompletion failed")
	}

	span.SetAttributes(
		attribute.Int("llm.input_tokens", resp.Usage.InputTokens),
		attribute.Int("llm.output_tokens", resp.Usage.OutputTokens),
		attribute.String("llm.stop_reason", string(resp.StopReason)),
	)

	log.Info("Anthropic:CreateMessages response", "resp", resp)
	usage := api.LLMUsage{
		InputTokens:  resp.Usage.InputTokens,
		OutputTokens: resp.Usage.OutputTokens,
		Model:        c.config.GetModel(),
		Provider:     string(api.ModelProviderAnthropic),
	}
	logs.LogLLMUsage(ctx, usage)

	blocks, err := c.parseResponse(ctx, &resp)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to parse response")
	}
	return blocks, nil
}

func (c *Completer) parseResponse(ctx context.Context, resp *anthropic.MessagesResponse) ([]*v1alpha1.Block, error) {
	log := logs.FromContext(ctx)
	allBlocks := make([]*v1alpha1.Block, 0, 10)
	for _, block := range resp.Content {
		if block.Type != anthropic.MessagesContentTypeText {
			log.Info("Skipping non text block in Anthropic Resposne", "type", block.Type)
			continue
		}

		blocks, err := docs.MarkdownToBlocks(*block.Text)
		if err != nil {
			log.Error(err, "Failed to parse markdown to blocks", "markdown", *block.Text)
			b := &v1alpha1.Block{
				Kind:     v1alpha1.BlockKind_MARKUP,
				Contents: *block.Text,
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
