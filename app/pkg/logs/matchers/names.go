// Package matchers contains matchers to detect if a log entry corresponds to a particular type of event.
// These function names are used to delegate the log message to the correct post processor.
// Correctness is ensured via the unittest which uses reflection to ensure the function names are correct.
package matchers

import "strings"

const (
	OAIComplete       = "github.com/jlewi/foyle/app/pkg/oaiapi.(*Completer).Complete"
	AnthropicComplete = "github.com/jlewi/foyle/app/pkg/anthropic.(*Completer).Complete"
	LogEvents         = "github.com/jlewi/foyle/app/pkg/agent.(*Agent).LogEvents"
	StreamGenerate    = "github.com/jlewi/foyle/app/pkg/agent.(*Agent).StreamGenerate"

	RequestField = "request"
	// ResponseField is the field storing the response of the LLM.
	// TODO(jeremy): The use of the abbreviation resp is inconsistent with the name of the field request but its what
	// we used.
	ResponseField = "resp"
)

type Matcher func(name string) bool

func IsOAIComplete(name string) bool {
	return strings.HasPrefix(name, OAIComplete)
}

func IsAnthropicComplete(name string) bool {
	return strings.HasPrefix(name, AnthropicComplete)
}

func IsLogEvent(fname string) bool {
	// We need to use HasPrefix because the logging statement is nested inside an anonymous function so there
	// will be a suffix like "func1"
	return strings.HasPrefix(fname, LogEvents)
}

func IsLLMUsage(fname string) bool {
	return strings.HasPrefix(fname, "github.com/jlewi/foyle/app/pkg/logs.LogLLMUsage")
}

func IsGenerate(fname string) bool {
	return strings.HasPrefix(fname, "github.com/jlewi/foyle/app/pkg/agent.(*Agent).Generate")
}

func IsStreamGenerate(fname string) bool {
	return strings.HasPrefix(fname, StreamGenerate)
}
