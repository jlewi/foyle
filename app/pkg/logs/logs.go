package logs

import (
	"context"
	"encoding/json"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

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

// ZapPB is a helper function to log a protobuf message as a field in a zap logger.
// See: https://stackoverflow.com/questions/68411821/correctly-log-protobuf-messages-as-unescaped-json-with-zap-logger
func ZapPB(name string, m proto.Message) zap.Field {
	b, err := protojson.Marshal(m)
	if err != nil {
		return zap.Error(err)
	}
	return zap.Any(name, json.RawMessage(b))
}
