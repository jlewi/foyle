package analyze

import (
	"fmt"
	"os"
	"testing"

	"github.com/liushuangls/go-anthropic/v2"
	"github.com/pkg/browser"
	"google.golang.org/protobuf/proto"
)

func TestRenderAnthropicRequest(t *testing.T) {
	type testCase struct {
		name    string
		request *anthropic.MessagesRequest
	}

	tests := []testCase{
		{
			name: "basic",
			request: &anthropic.MessagesRequest{
				Model:       "test",
				MaxTokens:   10,
				Temperature: proto.Float32(0.5),
				System:      "This is the system message",
				Messages: []anthropic.Message{
					{
						Role: "User",
						Content: []anthropic.MessageContent{
							{
								Text: proto.String("# md heading\n * item 1 \n * item 2"),
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := renderAnthropicRequest(test.request)
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
