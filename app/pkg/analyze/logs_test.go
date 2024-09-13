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
			result, err := readLLMLog(context.Background(), c.traceId, fullPath)
			if err != nil {
				t.Errorf("Failed to read Anthropic request: %v", err)
			}
			if result == nil {
				t.Fatalf("Request should not be nil")
			}
			if result.RequestHtml == "" {
				t.Errorf("Request should not be nil")
			}
			if result.ResponseHtml == "" {
				t.Errorf("Response should not be nil")
			}
			if result.RequestJson == "" {
				t.Errorf("Request should not be nil")
			}
			if result.ResponseJson == "" {
				t.Errorf("Response should not be nil")
			}
			t.Fatalf("TODO need to update the code to verify response is actually read since its on a separate line.")
		})
	}
}
