package anthropic

import (
	"context"
	"os"
	"testing"

	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/docs"
	"github.com/liushuangls/go-anthropic/v2"
)

func TestAnthropicCompleter(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("TestAnthropicCompleter is a manual test that is skipped in CICD")
	}

	if err := config.InitViper(nil); err != nil {
		t.Fatalf("Failed to initialize Viper: %v", err)
	}

	cfg := config.GetConfig()
	cfg.Agent.Model = anthropic.ModelClaude3Dot5Sonnet20240620
	client, err := NewClient(*cfg)
	if err != nil {
		t.Fatalf("Failed to create OpenAI client: %v", err)
	}

	completer, err := NewCompleter(*cfg, client)
	if err != nil {
		t.Fatalf("Failed to create Completer: %v", err)
	}
	blocks, err := completer.Complete(context.Background(), "You are a helpful assistant.", "Use gcloud to list GKE clusters")
	if err != nil {
		t.Fatalf("Failed to complete: %v", err)
	}

	t.Logf("Blocks: %+v", docs.BlocksToMarkdown(blocks))
}
