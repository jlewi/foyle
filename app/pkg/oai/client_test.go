package oai

import (
	"os"
	"testing"

	"github.com/jlewi/foyle/app/pkg/config"
)

func Test_BuildAzureAIConfig(t *testing.T) {
	f, err := os.CreateTemp("", "key.txt")
	if err != nil {
		t.Fatalf("Error creating temp file: %v", err)
	}
	if _, err := f.WriteString("somekey"); err != nil {
		t.Fatalf("Error writing to temp file: %v", err)
	}

	cfg := &config.Config{
		AzureOpenAI: &config.AzureOpenAIConfig{
			APIKeyFile: f.Name(),
			BaseURL:    "https://someurl.com",
			Deployments: []config.AzureDeployment{
				{
					Model:      DefaultModel,
					Deployment: "somedeployment",
				},
			},
		},
	}

	if err := f.Close(); err != nil {
		t.Fatalf("Error closing temp file: %v", err)
	}
	defer os.Remove(f.Name())

	clientConfig, err := buildAzureConfig(*cfg)
	if err != nil {
		t.Fatalf("Error building Azure config: %v", err)
	}

	if clientConfig.BaseURL != "https://someurl.com" {
		t.Fatalf("Expected BaseURL to be https://someurl.com but got %v", clientConfig.BaseURL)
	}
}
