package analyze

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jlewi/foyle/app/api"

	"github.com/liushuangls/go-anthropic/v2"
	"github.com/pkg/browser"
	"google.golang.org/protobuf/proto"
)

func TestRenderAnthropicRequest(t *testing.T) {
	type testCase struct {
		name     string
		fname    string
		provider api.ModelProvider
	}

	tests := []testCase{
		{
			name:     "openai",
			fname:    "openai_request.json",
			provider: api.ModelProviderOpenAI,
		},
		{
			name:     "anthropic",
			fname:    "anthropic_request.json",
			provider: api.ModelProviderAnthropic,
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

			result, err := RenderRequestHTML(string(data), test.provider)
			if err != nil {
				t.Fatalf("Failed to render request: %+v", err)
			}
			if result == "" {
				t.Fatalf("Request should not be empty")
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
			jsonData, err := json.Marshal(test.resp)
			if err != nil {
				t.Fatalf("Failed to unmarshal request: %v", err)
			}

			result, err := RenderResponseHTML(string(jsonData), api.ModelProviderAnthropic)
			if err != nil {
				t.Errorf("Failed to render response: %v", err)
			}
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

func Test_escapePrompt(t *testing.T) {
	type testCase struct {
		name     string
		prompt   string
		expected string
	}

	cases := []testCase{
		{
			name: "basic",
			prompt: `This is the input
<example>
some example
</example>
`,
			expected: `This is the input
&lt;example&gt;
some example
&lt;/example&gt;
`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := escapePromptTags(c.prompt)
			if result != c.expected {
				t.Errorf("Expected %s; got %s", c.expected, result)
			}
		})
	}
}
