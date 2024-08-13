package analyze

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestReadAnthropicRequest(t *testing.T) {
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
			result, err := readAnthropicRequest(context.Background(), c.traceId, fullPath)
			if err != nil {
				t.Errorf("Failed to read Anthropic request: %v", err)
			}
			if result == nil {
				t.Errorf("Request should not be nil")
			}
		})
	}
}
