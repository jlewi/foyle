package replicate

import (
	"strings"

	"github.com/go-logr/zapr"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/monogo/files"
	"github.com/pkg/errors"
	repGo "github.com/replicate/replicate-go"
	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

const (
	// ChatProxyBaseURL is the URL of the chat proxy for LLAMA on Replicate
	ChatProxyBaseURL = "https://openai-proxy.replicate.com/v1"
)

// TODO(jeremy): Should we implement an HTTP client with retries for the Replicate client?

func NewClient(cfg config.Config) (*repGo.Client, error) {
	if cfg.Replicate == nil {
		return nil, errors.New("Replicate config is nil; You must configure the replicate model provider to create a replicate client")
	}
	token, err := files.Read(cfg.Replicate.APIKeyFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to read Replicate API token from file %s", cfg.Replicate.APIKeyFile)
	}
	r8, err := repGo.NewClient(repGo.WithToken(strings.TrimSpace(string(token))))
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create Replicate client")
	}
	return r8, nil
}

// NewChatClient helper function to create a new OpenAI client from  a config
func NewChatClient(cfg config.Config) (*openai.Client, error) {
	if cfg.Replicate == nil {
		return nil, errors.New("Replicate config is nil; You must configure the replicate model provider to create a replicate client")
	}
	log := zapr.NewLogger(zap.L())
	// ************************************************************************
	// Setup middleware
	// ************************************************************************

	// Handle retryable errors
	// To handle retryable errors we use hashi corp's retryable client. This client will automatically retry on
	// retryable errors like 429; rate limiting
	retryClient := retryablehttp.NewClient()
	httpClient := retryClient.StandardClient()

	var clientConfig openai.ClientConfig
	log.Info("Configuring Replicate LLAMA chat proxy client")

	apiKey := ""
	if cfg.Replicate.APIKeyFile != "" {
		var err error
		apiBytes, err := files.Read(cfg.Replicate.APIKeyFile)
		if err != nil {
			return nil, err
		}
		apiKey = strings.TrimSpace(string(apiBytes))
	}

	if apiKey == "" {
		return nil, errors.New("APIKeyFile is required")
	}
	clientConfig = openai.DefaultConfig(apiKey)

	log.Info("Using Replicate chat proxy baseURL", "baseURL", ChatProxyBaseURL)
	clientConfig.BaseURL = ChatProxyBaseURL

	clientConfig.HTTPClient = httpClient
	client := openai.NewClientWithConfig(clientConfig)

	return client, nil
}
