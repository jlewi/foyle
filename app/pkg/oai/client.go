package oai

import (
	"net/url"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"

	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/hydros/pkg/files"
	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
)

const (

	// AzureOpenAIVersion is the version of the Azure OpenAI API to use.
	// For a list of versions see:
	// https://learn.microsoft.com/en-us/azure/ai-services/openai/reference#chat-completions
	AzureOpenAIVersion = "2024-02-01"
)

// NewClient helper function to create a new OpenAI client from  a config
func NewClient(cfg config.Config) (*openai.Client, error) {
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

	var clientConfig openai.ClientConfig
	if cfg.AzureOpenAI != nil {
		var clientErr error
		clientConfig, clientErr = buildAzureConfig(cfg)

		if clientErr != nil {
			return nil, clientErr
		}
	} else {
		log.Info("Configuring OpenAI client")
		if cfg.OpenAI == nil {
			return nil, errors.New("OpenAI configuration is required")
		}

		apiKey := ""
		if cfg.OpenAI.APIKeyFile != "" {
			var err error
			apiKey, err = readAPIKey(cfg.OpenAI.APIKeyFile)
			if err != nil {
				return nil, err
			}
		}
		// If baseURL is customized then we could be using a custom endpoint that may not require an API key
		if apiKey == "" && cfg.OpenAI.BaseURL == "" {
			return nil, errors.New("OpenAI APIKeyFile is required when using OpenAI")
		}
		clientConfig = openai.DefaultConfig(apiKey)
		if cfg.OpenAI.BaseURL != "" {
			log.Info("Using custom OpenAI BaseURL", "baseURL", cfg.OpenAI.BaseURL)
			clientConfig.BaseURL = cfg.OpenAI.BaseURL
		}
	}
	clientConfig.HTTPClient = httpClient
	client := openai.NewClientWithConfig(clientConfig)

	return client, nil
}

// buildAzureConfig helper function to create a new Azure OpenAI client config
func buildAzureConfig(cfg config.Config) (openai.ClientConfig, error) {
	apiKey, err := readAPIKey(cfg.AzureOpenAI.APIKeyFile)
	if err != nil {
		return openai.ClientConfig{}, err
	}
	u, err := url.Parse(cfg.AzureOpenAI.BaseURL)
	if err != nil {
		return openai.ClientConfig{}, errors.Wrapf(err, "could not parse Azure OpenAI BaseURL: %v", cfg.AzureOpenAI.BaseURL)
	}

	if u.Scheme != "https" {
		return openai.ClientConfig{}, errors.Errorf("Azure BaseURL %s is not valid; it must use the scheme https", cfg.AzureOpenAI.BaseURL)
	}

	// Check that all required models are deployed
	required := map[string]bool{
		config.DefaultModel: true,
	}

	for _, d := range cfg.AzureOpenAI.Deployments {
		delete(required, d.Model)
	}

	if len(required) > 0 {
		models := make([]string, 0, len(required))
		for m := range required {
			models = append(models, m)
		}
		return openai.ClientConfig{}, errors.Errorf("Missing Azure deployments for for OpenAI models %v; update AzureOpenAIConfig.deployments in your configuration to specify deployments for these models ", strings.Join(models, ", "))
	}
	log := zapr.NewLogger(zap.L())
	log.Info("Configuring Azure OpenAI", "baseURL", cfg.AzureOpenAI.BaseURL, "deployments", cfg.AzureOpenAI.Deployments)
	clientConfig := openai.DefaultAzureConfig(apiKey, cfg.AzureOpenAI.BaseURL)
	clientConfig.APIVersion = AzureOpenAIVersion
	mapper := AzureModelMapper{
		modelToDeployment: make(map[string]string),
	}
	for _, m := range cfg.AzureOpenAI.Deployments {
		mapper.modelToDeployment[m.Model] = m.Deployment
	}
	clientConfig.AzureModelMapperFunc = mapper.Map

	return clientConfig, nil
}

// AzureModelMapper maps OpenAI models to Azure deployments
type AzureModelMapper struct {
	modelToDeployment map[string]string
}

// Map maps an OpenAI model to an Azure deployment
func (m AzureModelMapper) Map(model string) string {
	log := zapr.NewLogger(zap.L())
	deployment, ok := m.modelToDeployment[model]
	if !ok {
		log.Error(errors.Errorf("No AzureAI deployment found for model %v", model), "missing deployment", "model", model)
		return "missing-deployment"
	}
	return deployment
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
