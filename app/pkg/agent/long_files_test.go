package agent

import (
	"context"
	"fmt"
	"github.com/jlewi/foyle/app/pkg/config"
	"github.com/jlewi/foyle/app/pkg/oai"
	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
	"github.com/sergi/go-diff/diffmatchpatch"
	"go.uber.org/zap"
	"os"
	"strings"
	"testing"
)

func run() error {
	inputFile := "/Users/jlewi/tmp/clusters.tf"

	if err := config.InitViper(nil); err != nil {
		return errors.Wrapf(err, "Error initializing viper")
	}

	lConfig := zap.NewDevelopmentConfig()
	l, err := lConfig.Build()
	if err != nil {
		return errors.Wrapf(err, "Error creating logger")
	}

	zap.ReplaceGlobals(l)

	cfg := config.GetConfig()

	client, err := oai.NewClient(*cfg)

	if err != nil {
		return errors.Wrapf(err, "Error creating OpenAI client")
	}

	segment, err := ReadFileSegment(inputFile, 5570, 5594)
	if err != nil {
		return errors.Wrapf(err, "Error reading file segment")
	}

	var sb strings.Builder
	input := editPromptInput{
		Changes: "Change the cluster u35 to use the region spaincentral. Add the label owner=foo to the cluster",
		Text:    segment.Text,
	}

	if err := editTemplate.Execute(&sb, input); err != nil {
		return errors.Wrapf(err, "Failed to execute prompt template")
	}

	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser,
			Content: sb.String(),
		},
	}
	request := openai.ChatCompletionRequest{
		Model:       openai.GPT4oMini,
		Messages:    messages,
		MaxTokens:   8000,
		Temperature: temperature,
	}

	resp, err := client.CreateChatCompletion(context.Background(), request)
	if err != nil {
		return errors.Wrapf(err, "Error calling OpenAI")
	}

	if len(resp.Choices) != 1 {
		return errors.Errorf("Expected 1 choice but got %v", len(resp.Choices))
	}

	output := resp.Choices[0].Message.Content
	fmt.Printf("ChatGPT Output:\n%v\n", output)

	// Compute a diff between the original text and the output
	dmp := diffmatchpatch.New()

	//diffs := dmp.DiffMain(segment.Text, output, false)
	//
	//// Create a patch
	//patches := dmp.PatchMake(segment.Text, diffs)

	patches := dmp.PatchMake(segment.Text, output)

	//// Convert the diff to unified format
	//patch, err := diff.ToUnified(ud, unifiedDiff)
	//if err != nil {
	//	fmt.Println("Error generating unified diff:", err)
	//	return
	//}
	for _, p := range patches {
		fmt.Printf("Patch:\n%v\n", p)
	}
	return nil
}

func Test_LongFiles(t *testing.T) {
	// The purpose of this test is to experiment with different patterns for modifying very long (e.g. 8K-10K lines)
	// files. Such as Terraform files
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("Test is skipped in GitHub actions")
	}

	if err := run(); err != nil {
		t.Fatalf("Error running test: %v", err)
	}
}
