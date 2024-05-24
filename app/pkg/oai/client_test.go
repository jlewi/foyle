package oai

import (
	"context"
	"os"
	"testing"

	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/sashabaranov/go-openai"
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
					Model:      config.DefaultModel,
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

func Test_Ollama(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("Test_Ollama is a manual test that is skipped in CICD")
	}
	clientCfg := openai.DefaultConfig("")
	clientCfg.BaseURL = "http://localhost:11434/v1"
	client := openai.NewClientWithConfig(clientCfg)

	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem,
			Content: "You are a helpful assistant.",
		},
		{Role: openai.ChatMessageRoleUser,
			Content: "hello",
		},
	}

	request := openai.ChatCompletionRequest{
		Model:       "llama2",
		Messages:    messages,
		MaxTokens:   2000,
		Temperature: 0.9,
	}

	resp, err := client.CreateChatCompletion(context.Background(), request)
	if err != nil {
		t.Fatalf("Failed to create chat completion: %v", err)
	}

	t.Logf("Response: %+v", resp)
}

func Test_OpenAIErrors(t *testing.T) {
	// We want to test how OpenAI errors are processed so that we ensure we surface actionable error messages.
	// E.g. if the API key is invalid, we want to surface that error message to the user.
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("Test_OpenAIErrors is a manual test that is skipped in CICD")
	}

	if err := config.InitViper(nil); err != nil {
		t.Fatalf("Failed to initialize viper: %v", err)
	}

	cfg := config.GetConfig()
	cfg.AzureOpenAI = nil

	//  Use a key file that should be deprecated
	cfg.OpenAI.APIKeyFile = "/Users/jlewi/secrets/openapi.api.key.20230207"

	client, err := NewClient(*cfg)
	if err != nil {
		t.Fatalf("Failed to create OpenAI client: %v", err)
	}

	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem,
			Content: "You are a helpful assistant.",
		},
		{Role: openai.ChatMessageRoleUser,
			Content: "hello",
		},
	}

	request := openai.ChatCompletionRequest{
		Model:       "llama2",
		Messages:    messages,
		MaxTokens:   2000,
		Temperature: 0.9,
	}

	resp, err := client.CreateChatCompletion(context.Background(), request)
	if err != nil {
		t.Fatalf("Failed to create chat completion: %v", err)
	}

	t.Logf("Response: %+v", resp)
}
