package analyze

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/liushuangls/go-anthropic/v2"
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
			resp := &anthropic.MessagesResponse{}
			if err := json.Unmarshal([]byte(result.ResponseJson), resp); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if resp.Model == "" {
				t.Errorf("Model should be set")
			}
		})
	}
}
