package oai

import (
	"strings"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/hydros/pkg/files"
	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
)

// NewClient helper function to create a new OpenAI client from  a config
func NewClient(cfg config.Config) (*openai.Client, error) {
	if cfg.OpenAI.APIKeyFile == "" {
		return nil, errors.New("OpenAI APIKeyFile is required")
	}
	apiKeyBytes, err := files.Read(cfg.OpenAI.APIKeyFile)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read OpenAI APIKeyFile: %v", cfg.OpenAI.APIKeyFile)
	}
	// make sure there is no leading or trailing whitespace
	apiKey := strings.TrimSpace(string(apiKeyBytes))

	// ************************************************************************
	// Setup middleware
	// ************************************************************************

	// Handle retryable errors
	// To handle retryable errors we use hashi corp's retryable client. This client will automatically retry on
	// retryable errors like 429; rate limiting
	retryClient := retryablehttp.NewClient()
	httpClient := retryClient.StandardClient()

	clientCfg := openai.DefaultConfig(apiKey)
	clientCfg.HTTPClient = httpClient
	client := openai.NewClientWithConfig(clientCfg)

	return client, nil
}
