package anthropic

import (
	"context"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/docs"
	"github.com/jlewi/foyle/app/pkg/llms"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/liushuangls/go-anthropic/v2"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"net/http"
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
	log := logs.FromContext(ctx)
	// See: https://docs.anthropic.com/en/api/messages
	// Claude doesn't have a system prompt.
	// First message must also be a user message.

	messages := []anthropic.Message{
		{Role: anthropic.RoleUser,
			// TODO(jeremy): We put the systemprompt and message as two separate content blocks. This was
			// just the most direct result of porting the OpenAI code. We might want to tweak that.
			Content: []anthropic.MessageContent{
				{Type: anthropic.MessagesContentTypeText,
					Text: proto.String(systemPrompt),
				},
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
	}

	log.Info("Anthropic:CreateMessages", "request", request)
	resp, err := c.client.CreateMessages(ctx, request)

	if err != nil {
		// https://docs.anthropic.com/en/api/errors
		aErr, ok := err.(*anthropic.RequestError)
		// 413 means context length exceeded.
		if ok && aErr.StatusCode == http.StatusRequestEntityTooLarge {
			return nil, llms.ContextLengthExceededError{Cause: err}
		}
		// TODO(jeremy): Should we surface the error to the user as blocks in the notebook
		return nil, errors.Wrapf(err, "CreateChatCompletion failed")
	}

	log.Info("Anthropic:CreateMessages response", "resp", resp)

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
