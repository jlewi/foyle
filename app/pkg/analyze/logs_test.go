package analyze

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestReadAnthropicLog(t *testing.T) {
	type testCase struct {
		name    string
		logFile string
		traceId string
	}
	cases := []testCase{
		{
			name:    "basic",
			logFile: "anthropic_logs.jsonl",
			traceId: "8d63ff91499d2c09be43be8a94d80d43",
		},
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Errorf("Failed to get current working directory")
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fullPath := filepath.Join(cwd, "test_data", c.logFile)
			result, err := readAnthropicLog(context.Background(), c.traceId, fullPath)
			if err != nil {
				t.Errorf("Failed to read Anthropic request: %v", err)
			}
			if result == nil {
				t.Fatalf("Request should not be nil")
			}
			if result.Request == nil {
				t.Errorf("Request should not be nil")
			}
			if result.Response == nil {
				t.Errorf("Response should not be nil")
			} else {
				if result.Response.Model == "" {
					t.Errorf("Model should not be empty")
				}
			}
		})
	}
}
