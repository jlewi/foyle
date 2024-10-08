package api

import (
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
)

// LLMUsage defines a standardized structure for LLM usage that can be independent of the model used.
type LLMUsage struct {
	// The number of tokens used.
	InputTokens int `json:"inputTokens"`
	// The number of tokens generated.
	OutputTokens int `json:"outputTokens"`
	// Model used
	Model string `json:"model"`
	// Provider used
	Provider string `json:"provider"`
}

func ModelProviderProtoToAPI(provider v1alpha1.ModelProvider) ModelProvider {
	switch provider {
	case v1alpha1.ModelProvider_OPEN_AI:
		return ModelProviderOpenAI
	case v1alpha1.ModelProvider_ANTHROPIC:
		return ModelProviderAnthropic
	default:
		return ModelProviderUnknown
	}
}
