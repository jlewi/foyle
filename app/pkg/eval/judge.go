package eval

import (
	"context"
	_ "embed"
	"encoding/json"
	"strings"
	"text/template"

	"github.com/jlewi/foyle/app/pkg/docs"
	"github.com/jlewi/foyle/app/pkg/runme/converters"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	parserv1 "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/parser/v1"
)

//go:embed judge_prompt.tmpl
var promptTemplateString string

var (
	promptTemplate = template.Must(template.New("judge_prompt").Parse(promptTemplateString))
)

const (
	temperature = 0.9
)

type promptArgs struct {
	Expected string
	Actual   string
}

func NewJudge(client *openai.Client) (*Judge, error) {
	return &Judge{
		client: client,
	}, nil
}

// Judge is an LLM judge
type Judge struct {
	client *openai.Client
}

func (j *Judge) Score(ctx context.Context, result *v1alpha1.EvalResult) error {
	ctx, span := tracer().Start(ctx, "(*Judge).Score")
	defer span.End()

	if len(result.GetExample().ExpectedCells) != 1 {
		return errors.New("expected a single expected cell")
	}

	// We don't check if actualCells is 1 because if its empty then the program is wrong and the judge
	// should hopefully detect that.

	// Convert the cells to markdown
	expectedNB := &parserv1.Notebook{
		Cells: result.GetExample().ExpectedCells,
	}

	actualNB := &parserv1.Notebook{
		Cells: result.GetActualCells(),
	}

	expectedDoc, err := converters.NotebookToDoc(expectedNB)
	if err != nil {
		return errors.Wrap(err, "Failed to convert expected cells to doc")
	}

	actualDoc, err := converters.NotebookToDoc(actualNB)
	if err != nil {
		return errors.Wrap(err, "Failed to convert actual cells to doc")
	}

	expectedMD := docs.DocToMarkdown(expectedDoc)
	actualMD := docs.DocToMarkdown(actualDoc)

	args := promptArgs{
		Expected: expectedMD,
		Actual:   actualMD,
	}

	var sb strings.Builder
	if err := promptTemplate.Execute(&sb, args); err != nil {
		return errors.Wrapf(err, "Failed to execute prompt template")
	}

	messages := []openai.ChatCompletionMessage{
		//{Role: openai.ChatMessageRoleSystem,
		//	Content: systemPrompt,
		//},
		{Role: openai.ChatMessageRoleUser,
			Content: sb.String(),
		},
	}

	// TODO(jeremy): Use ResponseFormat to enforce JSON output
	// https://platform.openai.com/docs/guides/structured-outputs/how-to-use?context=without_parse
	request := openai.ChatCompletionRequest{
		// TODO(jeremy): Should we use gpt4 mini
		Model:       openai.GPT4o20240806,
		Messages:    messages,
		MaxTokens:   2000,
		Temperature: temperature,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name: "JudgeOutput",
				Schema: &jsonschema.Definition{
					Type:        jsonschema.Object,
					Description: "",
					Enum:        nil,
					// TODO(
					Properties: map[string]jsonschema.Definition{
						"equivalent": {
							Type: jsonschema.Boolean,
						},
						"reason": {
							Type: jsonschema.String,
						},
					},
					Required: []string{"equivalent", "reason"},
				},
			},
		},
	}

	resp, err := j.client.CreateChatCompletion(ctx, request)
	if err != nil {
		return err
	}

	if len(resp.Choices) == 0 {
		return errors.New("No choices in response")
	}

	choice := resp.Choices[0]
	output := &JudgeOutput{}
	if err := json.Unmarshal([]byte(choice.Message.Content), output); err != nil {
		return errors.Wrap(err, "Failed to unmarshal output")
	}

	if output.Equivalent {
		result.CellsMatchResult = v1alpha1.CellsMatchResult_MATCH
	} else {
		result.CellsMatchResult = v1alpha1.CellsMatchResult_MISMATCH
	}

	result.JudgeExplanation = output.Reason

	return nil
}

// JudgeOutput is the JSON output we expect the judge to emit
type JudgeOutput struct {
	Equivalent bool   `json:"equivalent"`
	Reason     string `json:"reason"`
}
