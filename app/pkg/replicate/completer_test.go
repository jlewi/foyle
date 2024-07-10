package replicate

import (
	"context"
	"os"
	"testing"

	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/docs"
)

func Test_ReplicateCompleter(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("Test is skipped in GitHub actions")
	}

	cfg := config.Config{
		Replicate: &config.ReplicateConfig{
			APIKeyFile: "/Users/jlewi/replicate/secrets/apikey",
		},
	}

	client, err := NewChatClient(cfg)
	if err != nil {
		t.Fatalf("Error creating client; %v", err)
	}

	completer, err := NewCompleter(cfg, client)
	if err != nil {
		t.Fatalf("Error creating completer; %v", err)
	}

	systemPrompt := "You are a cloud expert. Help the user with the following task"
	message := "Use gcloud to list buckets"
	result, err := completer.Complete(context.Background(), systemPrompt, message)
	if err != nil {
		t.Fatalf("Error completing text; %v", err)
	}

	if len(result) == 0 {
		t.Fatalf("Complete returned no results")
	}

	t.Logf("Result:\n%v", docs.BlocksToMarkdown(result))
}
