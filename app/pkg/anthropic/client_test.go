package anthropic

import (
	"context"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/liushuangls/go-anthropic/v2"
	"google.golang.org/protobuf/proto"
	"os"
	"testing"
)

func Test_AnthropicClient(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("TestAnthropicClient is a manual test that is skipped in CICD")
	}

	// N.B since this test isn't using a valid endpoint we don't expect it to succceed
	// But hopefully the error message will let us confirm that the baseURL is being used

	// Test non standard base URL
	cfg := config.Config{
		Anthropic: &config.AnthropicConfig{
			APIKeyFile: "/Users/jlewi/secrets/anthropic.key",
			BaseURL:    "https://localhost:8844",
		},
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create Anthropic client: %v", err)
	}

	message := "Say hello a funny way"
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
		Model:       cfg.GetModel(),
		Messages:    messages,
		MaxTokens:   2000,
		Temperature: proto.Float32(temperature),
		System:      "You are a helper",
	}

	client.CreateMessages(context.Background(), request)
}
