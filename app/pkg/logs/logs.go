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
		return zapr.NewLogger(zap.L())
	}
	return l
}

func ZapFromLogr(log logr.Logger) *zap.Logger {
	u, ok := log.GetSink().(zapr.Underlier)
	if !ok {
		return zap.L()
	}
	return u.GetUnderlying()
}
