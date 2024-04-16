package analyze

import (
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"time"
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
	v, ok := (*L)["traceId"]
	if !ok {
		return ""
	}
	if v, ok := v.(string); ok {
		return v
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

type TraceType string

const (
	GenerateTraceType TraceType = "Generate"
	ExecuteTraceType  TraceType = "Execute"
)

type Trace interface {
	// ID is the id of this trace
	ID() string
	Type() TraceType
}

// GenerateTrace is the trace of a generation request.
type GenerateTrace struct {
	// ID is the id of this trace
	TraceID   string                     `json:"traceId"`
	StartTime time.Time                  `json:"startTime"`
	EndTime   time.Time                  `json:"endTime"`
	Request   *v1alpha1.GenerateRequest  `json:"request"`
	Response  *v1alpha1.GenerateResponse `json:"response"`
}

func (g *GenerateTrace) ID() string {
	return g.TraceID
}

func (g *GenerateTrace) Type() TraceType {
	return GenerateTraceType
}

// ExecuteTrace is the trace of an execution request.
type ExecuteTrace struct {
	// ID is the id of this trace
	TraceID   string                    `json:"traceId"`
	StartTime time.Time                 `json:"startTime"`
	EndTime   time.Time                 `json:"endTime"`
	Request   *v1alpha1.ExecuteRequest  `json:"request"`
	Response  *v1alpha1.ExecuteResponse `json:"response"`
}

func (e *ExecuteTrace) ID() string {
	return e.TraceID
}

func (e *ExecuteTrace) Type() TraceType {
	return ExecuteTraceType
}

// BlockLog is the log of what happened to a block. It includes information about how a block was generated (if it
// was generated by the AI) and how it was executed if it was.
type BlockLog struct {
	// ID is the id of this block
	ID string `json:"id"`

	// GenTraceID is the trace ID of the generation request
	GenTraceID string `json:"genTraceId"`

	// ExecTraceIDs are the trace IDs of the execution requests
	ExecTraceIDs []string `json:"execTraceId"`

	// Doc is the doc that triggered the generated block
	Doc *v1alpha1.Doc `json:"doc"`
	// GeneratedBlock is the block generated by the AI
	GeneratedBlock *v1alpha1.Block `json:"generatedBlock"`
	// ExecutedBlock is the final block that was actually executed
	// nil if the block was not executed
	ExecutedBlock *v1alpha1.Block `json:"executedBlock"`

	// ExitCode is the exit code of the executed block
	ExitCode int `json:"exitCode"`
}
