// Package fnames contains constants defining the function names that are logged in certain log messages.
// These function names are used to delegate the log message to the correct post processor.
// Correctness is ensured via the unittest which uses reflection to ensure the function names are correct.
package fnames

const (
	LogEvents = "github.com/jlewi/foyle/app/pkg/agent.(*Agent).LogEvents"
)
