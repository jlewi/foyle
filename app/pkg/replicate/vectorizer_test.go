package replicate

import (
	"context"
	"os"
	"testing"

	"github.com/jlewi/foyle/app/pkg/config"
)

func TestVectorizer_Embed(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skip("Skipping test on GitHub Actions.")
	}

	cfg := config.Config{
		Replicate: &config.ReplicateConfig{
			APIKeyFile: "/Users/jlewi/replicate/secrets/apikey",
		},
	}

	client, err := NewClient(cfg)

	if err != nil {
		t.Fatalf("Error creating client; %v", err)
	}

	v, err := NewVectorizer(client)
	if err != nil {
		t.Fatalf("Error creating vectorizer; %v", err)
	}

	result, err := v.Embed(context.Background(), "Hello World")
	if err != nil {
		t.Fatalf("Error embedding text; %v", err)
	}

	if result == nil {
		t.Fatalf("Embed returned nil")
	}

	if result.Len() != v.Length() {
		t.Fatalf("Expected length of %d; got %d", v.Length(), result.Len())
	}
}
