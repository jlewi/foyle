package logs

import (
	"context"
	"encoding/json"

	"github.com/jlewi/foyle/app/api"

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

	// Level1Assertion Message denoting a level1 assertion
	Level1Assertion = "Level1Assert"
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

// ZapProto is a helper function to be able to log protos as JSON objects.
// We want protos to be logged using the proto json format so we can deserialize them from the logs.
// If you just log a proto with zap it will use the json serialization of the GoLang struct which will not match
// the proto json format. So we serialize the request to JSON and then deserialize it to a map so we can log it as a
// JSON object. A more efficient solution would be to use https://github.com/kazegusuri/go-proto-zap-marshaler
// to generate a custom zapcore.ObjectMarshaler implementation for each proto message.
func ZapProto(key string, pb proto.Message) zap.Field {
	log := NewLogger()
	reqObj := map[string]interface{}{}
	reqJSON, err := protojson.Marshal(pb)
	if err != nil {
		log.Error(err, "failed to marshal request")
		reqObj["error"] = err.Error()
		return zap.Any(key, reqObj)
	}

	if err := json.Unmarshal(reqJSON, &reqObj); err != nil {
		log.Error(err, "failed to unmarshal request")
		reqObj["error"] = err.Error()
	}

	return zap.Any(key, reqObj)
}

// LogLLMUsage logs the usage of the LLM model
// The purpose of this utility function is to create a standard log message independent of the model being used.
// This simplifies post-processing
func LogLLMUsage(ctx context.Context, usage api.LLMUsage) {
	log := FromContext(ctx)
	log.Info("LLM usage", "usage", usage)
}
