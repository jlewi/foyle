package api

import (
	"encoding/json"
	"time"

	"github.com/jlewi/foyle/app/pkg/logs/matchers"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

const (
	// TraceIDField is the field name for the trace ID used in Foyle logs
	TraceIDField = "traceId"

	// RunMeIDField is the field name for the trace id used in RunMe Logs
	RunMeIDField = "_id"
)

// LogEntry represents a log entry.
// We can't use a struct with well known types because when logging structured data we don't enforce a field
// name to be of a certain type. e.g. we could have
// log.Info("Generate Request", "request", &v1alpha1.GenerateRequest{})
// log.Info("Execute Request", "request", &v1alpha1.ExecuteRequest{})
// So the type of the "request" field is context dependent.
type LogEntry map[string]interface{}

func (L *LogEntry) Get(field string) (interface{}, bool) {
	v, ok := (*L)[field]
	return v, ok
}

func (L *LogEntry) GetBool(field string) (bool, bool) {
	v, ok := (*L)[field]
	if !ok {
		return false, false
	}

	f, ok := v.(bool)
	return f, ok
}

func (L *LogEntry) GetString(field string) (string, bool) {
	v, ok := (*L)[field]
	if !ok {
		return "", false
	}

	f, ok := v.(string)
	return f, ok
}

func (L *LogEntry) GetFloat64(field string) (float64, bool) {
	v, ok := (*L)[field]
	if !ok {
		return 0, false
	}

	f, ok := v.(float64)
	return f, ok
}

// GetProto gets the field as the specified proto message.
// Returns false if the field was not present or the field was not the supplied proto message.
func (L *LogEntry) GetProto(field string, msg proto.Message) bool {
	v, ok := (*L)[field]
	if !ok {
		return ok
	}
	obj, ok := v.(map[string]interface{})

	if !ok {
		return false
	}

	b, err := json.Marshal(obj)
	if err != nil {
		log := zapr.NewLogger(zap.L())
		log.Error(err, "Failed to marshal request")
		return false
	}
	if err := protojson.Unmarshal(b, msg); err != nil {
		log := zapr.NewLogger(zap.L())
		log.Error(err, "Failed to unmarshal request")
		return false

	}
	return true
}

// GetStruct gets the field as the specified GoStruct.
// Returns false if the field was not present or the field was not the supplied proto message.
func (L *LogEntry) GetStruct(field string, s interface{}) bool {
	v, ok := (*L)[field]
	if !ok {
		return false
	}
	obj, ok := v.(map[string]interface{})
	if !ok {
		return false
	}
	b, err := json.Marshal(obj)
	if err != nil {
		log := zapr.NewLogger(zap.L())
		log.Error(err, "Failed to marshal request")
		return false
	}

	if err := json.Unmarshal(b, s); err != nil {
		log := zapr.NewLogger(zap.L())
		log.Error(err, "Failed to unmarshal request")
		return false
	}

	return true
}

func (L *LogEntry) Request() []byte {
	// Different field names can be quest for the request.
	// Foyle uses "request" and RunMe uses "req"
	for _, field := range []string{"request", "req"} {
		v, ok := (*L)[field]
		if !ok {
			continue
		}
		if v, ok := v.(map[string]interface{}); ok {
			b, err := json.Marshal(v)
			if err != nil {
				log := zapr.NewLogger(zap.L())
				log.Error(err, "Failed to marshal request")
				return nil
			}
			return b
		}
	}
	return nil
}

// EvalMode returns value, ok. Where ok is true if the field was present and false otherwise.
// If the field was present value is the value of the field.
func (L *LogEntry) EvalMode() (bool, bool) {
	v, ok := (*L)["evalMode"]
	if !ok {
		return false, false
	}
	if val, ok := v.(bool); ok {
		return val, true
	}
	return false, false
}

func (L *LogEntry) Response() []byte {
	// We use "response" in some places (e.g. Foyle APIs) but "resp" in others (e.g. in Completers) when logging
	// LLM calls.
	for _, field := range []string{"response", "resp"} {
		v, ok := (*L)[field]
		if !ok {
			continue
		}
		if v, ok := v.(map[string]interface{}); ok {
			b, err := json.Marshal(v)
			if err != nil {
				log := zapr.NewLogger(zap.L())
				log.Error(err, "Failed to marshal response")
				return nil
			}
			return b
		}
	}
	return nil
}

func (L *LogEntry) Function() string {
	v, ok := (*L)["function"]
	if !ok {
		return ""
	}
	if v, ok := v.(string); ok {
		return v
	}
	return ""
}

func (L *LogEntry) Message() string {
	v, ok := (*L)["message"]
	if !ok {
		return ""
	}
	if v, ok := v.(string); ok {
		return v
	}
	return ""
}

func (L *LogEntry) TraceID() string {
	for _, field := range []string{TraceIDField, RunMeIDField} {
		v, ok := (*L)[field]
		if ok {
			if v, ok := v.(string); ok {
				return v
			}
		}
	}
	return ""
}

func (L *LogEntry) Time() time.Time {
	v, ok := (*L)["time"]
	if !ok {
		return time.Time{}
	}
	floatV, ok := v.(float64)
	if !ok {
		return time.Time{}
	}
	seconds := int64(floatV)
	fractional := floatV - float64(seconds)
	nanoseconds := int64(fractional * 1e9)

	timestamp := time.Unix(seconds, nanoseconds)
	return timestamp
}

// SetRequest sets the request field in the log entry.
// This is only intended for constructing log entries as part of testing
func SetRequest(e *LogEntry, req interface{}) error {
	b, err := json.Marshal(req)
	if err != nil {
		return err
	}

	o := make(map[string]interface{})
	if err := json.Unmarshal(b, &o); err != nil {
		return err
	}
	(*e)[matchers.RequestField] = o
	return nil
}

// SetResponse sets the response field in the log entry.
// This is only intended for constructing log entries as part of testing
func SetResponse(e *LogEntry, req interface{}) error {
	b, err := json.Marshal(req)
	if err != nil {
		return err
	}

	o := make(map[string]interface{})
	if err := json.Unmarshal(b, &o); err != nil {
		return err
	}
	(*e)[matchers.ResponseField] = o
	return nil
}
