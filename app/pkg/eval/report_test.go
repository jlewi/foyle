package eval

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/pkg/browser"
	"github.com/sashabaranov/go-openai"
	parserv1 "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/parser/v1"
)

func Test_buildReport(t *testing.T) {
	result := &v1alpha1.EvalResult{
		Example: &v1alpha1.EvalExample{
			Id:   "someid",
			Time: nil,
			FullContext: &v1alpha1.FullContext{
				Notebook:    nil,
				Selected:    5,
				NotebookUri: "file://somefile",
			},
			ExpectedCells: []*parserv1.Cell{
				{
					Kind:  parserv1.CellKind_CELL_KIND_CODE,
					Value: "expected cell value",
				},
			},
		},
		CellsMatchResult: v1alpha1.CellsMatchResult_MATCH,
		JudgeExplanation: `This is the judge's explanation
1. Item 1
1. Item 2 
`,
	}

	oaiRequest := openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "System message",
			},
			{
				Role: openai.ChatMessageRoleUser,
				Content: `# Some Markdown heading

<example>
some example
</example>

<input>
input
`,
			},
		},
	}

	requestJson, err := json.Marshal(oaiRequest)
	if err != nil {
		t.Fatalf("Error marshalling request: %v", err)
	}

	oaiResponse := &openai.ChatCompletionResponse{
		Choices: []openai.ChatCompletionChoice{
			{
				Message: openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleSystem,
					Content: "This is the response message",
				},
			},
		},
	}

	responseJson, err := json.Marshal(oaiResponse)
	if err != nil {
		t.Fatalf("Error marshalling response: %v", err)
	}

	trace := &logspb.Trace{
		Spans: []*logspb.Span{
			{
				Data: &logspb.Span_Llm{
					Llm: &logspb.LLMSpan{
						Provider:     v1alpha1.ModelProvider_OPEN_AI,
						RequestJson:  string(requestJson),
						ResponseJson: string(responseJson),
					},
				},
			},
		},
	}

	reportHtml, err := buildEvalReport(context.Background(), result, trace)

	if err != nil {
		t.Fatalf("Error building report: %v", err)
	}

	if os.Getenv("OPEN_IN_BROWSER") != "" {
		name := "/tmp/evalreport.html"
		if err := os.WriteFile(name, []byte(reportHtml), 0644); err != nil {
			t.Errorf("Failed to write file %s: %v", name, err)
		} else {
			if err := browser.OpenURL("file://" + name); err != nil {
				t.Errorf("Failed to open browser: %v", err)
			}
		}
	}

	t.Logf("Result HTML:\n%s", result)

}
