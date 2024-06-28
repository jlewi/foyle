package oai

import (
	"context"

	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/docs"
	"github.com/jlewi/foyle/app/pkg/llms"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
)

const (
	temperature = 0.9
)

func NewCompleter(cfg config.Config, client *openai.Client) (*Completer, error) {
	return &Completer{
		client: client,
		config: cfg,
	}, nil
}

// Completer is a wrapper around OpenAI that implements the Completer interface.
type Completer struct {
	client *openai.Client
	config config.Config
}

// Complete returns a ContextLengthExceededError if the context is too long
func (c *Completer) Complete(ctx context.Context, systemPrompt string, message string) ([]*v1alpha1.Block, error) {
	log := logs.FromContext(ctx)
	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
		{Role: openai.ChatMessageRoleUser,
			Content: message,
		},
	}
	request := openai.ChatCompletionRequest{
		Model:       c.config.GetModel(),
		Messages:    messages,
		MaxTokens:   2000,
		Temperature: temperature,
	}

	log.Info("OpenAI:CreateChatCompletion", "request", request)
	resp, err := c.client.CreateChatCompletion(ctx, request)

	if err != nil {
		if ErrorIs(err, ContextLengthExceededCode) {
			return nil, llms.ContextLengthExceededError{Cause: err}
		}
		// TODO(jeremy): Should we surface the error to the user as blocks in the notebook
		return nil, errors.Wrapf(err, "CreateChatCompletion failed")
	}

	log.Info("OpenAI:CreateChatCompletion response", "resp", resp)

	blocks, err := c.parseResponse(ctx, &resp)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to parse response")
	}
	return blocks, nil
}

func (c *Completer) parseResponse(ctx context.Context, resp *openai.ChatCompletionResponse) ([]*v1alpha1.Block, error) {
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
