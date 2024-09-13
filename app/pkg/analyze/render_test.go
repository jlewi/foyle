package analyze

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/liushuangls/go-anthropic/v2"
	"github.com/pkg/browser"
	"google.golang.org/protobuf/proto"
)

func TestRenderAnthropicRequest(t *testing.T) {
	type testCase struct {
		name  string
		fname string
	}

	tests := []testCase{
		{
			name:  "basic",
			fname: "anthropic_request.json",
		},
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	testDataDir := filepath.Join(cwd, "test_data")
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			fname := filepath.Join(testDataDir, test.fname)
			data, err := os.ReadFile(fname)
			if err != nil {
				t.Fatalf("Failed to read file %s: %v", fname, err)
			}

			req := &anthropic.MessagesRequest{}
			json.Unmarshal(data, req)

			result := renderAnthropicRequest(req)
			if result == "" {
				t.Errorf("Request should not be empty")
			}

			if os.Getenv("OPEN_IN_BROWSER") != "" {
				name := fmt.Sprintf("/tmp/%s.request.html", test.name)
				if err := os.WriteFile(name, []byte(result), 0644); err != nil {
					t.Errorf("Failed to write file %s: %v", name, err)
				} else {
					if err := browser.OpenURL("file://" + name); err != nil {
						t.Errorf("Failed to open browser: %v", err)
					}
				}
			}

			t.Logf("Result HTML:\n%s", result)
		})
	}
}

func TestRenderAnthropicResponse(t *testing.T) {
	type testCase struct {
		name string
		resp *anthropic.MessagesResponse
	}

	tests := []testCase{
		{
			name: "basic",
			resp: &anthropic.MessagesResponse{
				Model: "test",
				ID:    "test-id",
				Role:  "assistant",
				Usage: anthropic.MessagesUsage{
					InputTokens:  103,
					OutputTokens: 105,
				},
				StopReason:   anthropic.MessagesStopReasonStopSequence,
				StopSequence: "stop",
				Content: []anthropic.MessageContent{
					{
						Text: proto.String("This is the response message"),
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := renderAnthropicResponse(test.resp)
			if result == "" {
				t.Errorf("Result should not be empty")
			}

			if os.Getenv("OPEN_IN_BROWSER") != "" {
				name := fmt.Sprintf("/tmp/%s.request.html", test.name)
				if err := os.WriteFile(name, []byte(result), 0644); err != nil {
					t.Errorf("Failed to write file %s: %v", name, err)
				} else {
					if err := browser.OpenURL("file://" + name); err != nil {
						t.Errorf("Failed to open browser: %v", err)
					}
				}
			}

			t.Logf("Result HTML:\n%s", result)
		})
	}
}
