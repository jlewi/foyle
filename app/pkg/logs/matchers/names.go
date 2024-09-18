// Package matchers contains matchers to detect if a log entry corresponds to a particular type of event.
// These function names are used to delegate the log message to the correct post processor.
// Correctness is ensured via the unittest which uses reflection to ensure the function names are correct.
package matchers

import "strings"

const (
	OAIComplete    = "github.com/jlewi/foyle/app/pkg/oai.(*Completer).Complete"
	LogEvents      = "github.com/jlewi/foyle/app/pkg/agent.(*Agent).LogEvents"
	StreamGenerate = "github.com/jlewi/foyle/app/pkg/agent.(*Agent).StreamGenerate"
)

type Matcher func(name string) bool

func IsOAIComplete(name string) bool {
	return strings.HasPrefix(name, OAIComplete)
}

func IsLogEvent(fname string) bool {
	// We need to use HasPrefix because the logging statement is nested inside an anonymous function so there
	// will be a suffix like "func1"
	return strings.HasPrefix(fname, LogEvents)
}

func IsLLMUsage(fname string) bool {
	return strings.HasPrefix(fname, "github.com/jlewi/foyle/app/pkg/logs.LogLLMUsage")
}

func IsStreamGenerate(fname string) bool {
	return strings.HasPrefix(fname, StreamGenerate)
}
