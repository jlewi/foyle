package anthropic

import (
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/go-logr/zapr"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/monogo/files"
	"github.com/liushuangls/go-anthropic/v2"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// NewClient helper function to create a new OpenAI client from  a config
func NewClient(cfg config.Config) (*anthropic.Client, error) {
	log := zapr.NewLogger(zap.L())
	// ************************************************************************
	// Setup middleware
	// ************************************************************************

	// Handle retryable errors
	// To handle retryable errors we use hashi corp's retryable client. This client will automatically retry on
	// retryable errors like 429; rate limiting
	retryClient := retryablehttp.NewClient()
	httpClient := retryClient.StandardClient()

	if cfg.UseHoneycomb() {
		httpClient.Transport = otelhttp.NewTransport(httpClient.Transport)
	}

	if cfg.Anthropic == nil {
		return nil, errors.New("Anthropic config is nil; You must configure Anthropic to create an Anthropic client")
	}

	log.Info("Configuring Anthropic client")

	apiKey := ""
	if cfg.Anthropic.APIKeyFile == "" {
		return nil, errors.New("Anthropic APIKeyFile is required when using Anthropic")
	}

	var err error
	apiKey, err = readAPIKey(cfg.Anthropic.APIKeyFile)
	if err != nil {
		return nil, err
	}

	client := anthropic.NewClient(apiKey, anthropic.WithHTTPClient(httpClient))
	return client, nil
}

func readAPIKey(apiKeyFile string) (string, error) {
	if apiKeyFile == "" {
		return "", errors.New("APIKeyFile is required")
	}
	apiKeyBytes, err := files.Read(apiKeyFile)
	if err != nil {
		return "", errors.Wrapf(err, "could not read APIKeyFile: %v", apiKeyFile)
	}
	// make sure there is no leading or trailing whitespace
	apiKey := strings.TrimSpace(string(apiKeyBytes))
	return apiKey, nil
}
