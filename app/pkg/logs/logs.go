package logs

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

const (
	// Verbosity level constants for logr.
	// With logr verbosity is additive
	// e.g. log.V(1).Info() means verbosity = info verbosity + 1

	// Debug is for debug verbosity level
	Debug = 1
)

// FromContext returns a logr.Logger from the context or an instance of the global logger
func FromContext(ctx context.Context) logr.Logger {
	l, err := logr.FromContext(ctx)
	if err != nil {
		return NewLogger()
	}
	return l
}

func NewLogger() logr.Logger {
	// We need to AllowZapFields to ensure the protobuf message is logged correctly as a json object.
	// For that to work we need to do logr.Info("message", zap.Object("key", protoMessage))
	// Which means we are passing zap.Field to the logr interface.
	return zapr.NewLoggerWithOptions(zap.L(), zapr.AllowZapFields(true))
}
