package oai

import "github.com/sashabaranov/go-openai"

const (
	// ContextLengthExceededCode the error code returned by OpenAI to indicate the context length was exceeded
	ContextLengthExceededCode = "context_length_exceeded"
)

// ErrorIs checks if the error is an OpenAI error with the given code.
func ErrorIs(err error, oaiCode string) bool {
	apiErr, ok := err.(*openai.APIError)
	if !ok {
		return false
	}

	val, ok := apiErr.Code.(string)
	if !ok {
		return false
	}

	return val == oaiCode
}
