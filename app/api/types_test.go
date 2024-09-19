package api

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/go-logr/zapr"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
)

func Test_Deserialize(t *testing.T) {
	type testCase struct {
		name    string
		input   string
		message string
		time    time.Time
		traceID string
	}

	cases := []testCase{
		{
			name:    "simple",
			input:   `{"severity":"INFO","time":1713208269,"message":"hello","traceId":"1234","blockIds":["1","2"]}`,
			message: "hello",
			time:    time.Unix(1713208269, 0),
			traceID: "1234",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var actual LogEntry
			if err := json.Unmarshal([]byte(c.input), &actual); err != nil {
				t.Fatalf("failed to deserialize: %v", err)
			}

			if actual.Message() != c.message {
				t.Errorf("Expected message %v but got %v", c.message, actual.Message())

			}
			if actual.Time() != c.time {
				t.Errorf("Expected time %v but got %v", c.time, actual.Time())
			}
			if actual.TraceID() != c.traceID {
				t.Errorf("Expected traceID %v but got %v", c.traceID, actual.TraceID())
			}
		})
	}
}

func Test_GetStruct(t *testing.T) {
	type testCase struct {
		name     string
		input    string
		field    string
		expected interface{}
	}

	// Create a temp file to store the logs
	fname, err := os.CreateTemp("", "testGetStruct.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(fname.Name())
	if err := fname.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	// Configure a logger to log some data to generate test data
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{fname.Name()}
	zapL, err := config.Build()
	if err != nil {
		t.Fatalf("failed to build logger: %v", err)
	}
	log := zapr.NewLogger(zapL)
	usage := &LLMUsage{
		InputTokens:  10,
		OutputTokens: 20,
		Model:        "model",
		Provider:     "provider",
	}
	log.Info("LLMUsage", "usage", usage)
	if err := zapL.Sync(); err != nil {
		t.Fatalf("failed to sync logger: %v", err)
	}
	loggedBytes, err := os.ReadFile(fname.Name())
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	cases := []testCase{
		{
			name:     "simple",
			input:    string(loggedBytes),
			field:    "usage",
			expected: usage,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var actual LogEntry
			if err := json.Unmarshal([]byte(c.input), &actual); err != nil {
				t.Fatalf("failed to deserialize: %v", err)
			}

			actualUsage := &LLMUsage{}
			if ok := actual.GetStruct(c.field, actualUsage); !ok {
				t.Fatalf("failed to get struct")
			}

			if d := cmp.Diff(c.expected, actualUsage); d != "" {
				t.Errorf("unexpected diff:\n%v", d)
			}
		})
	}
}
