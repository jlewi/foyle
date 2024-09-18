package api

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
