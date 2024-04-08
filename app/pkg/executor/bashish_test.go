package executor

import (
	"strings"
	"testing"

	"github.com/go-cmd/cmd"
	"github.com/google/go-cmp/cmp"
)

func Test_BashishParser(t *testing.T) {
	type testCase struct {
		name     string
		lines    []string
		expected []Instruction
	}

	cases := []testCase{
		{
			name:  "basic",
			lines: []string{"kubectl get pods"},
			expected: []Instruction{
				{
					Command: cmd.NewCmd("kubectl", "get", "pods"),
				},
			},
		},
		{
			// This text mimics what you would get if you typed the command into a shell
			name:  "quoted",
			lines: []string{"echo \"some text\""},
			expected: []Instruction{
				{
					Command: cmd.NewCmd("echo", "some text"),
				},
			},
		},
		{
			name:  "simple-pipe",
			lines: []string{"ls -la | wc -l"},
			expected: []Instruction{
				{
					Command: cmd.NewCmd("ls", "-la"),
					Piped:   true,
				},
				{
					Command: cmd.NewCmd("wc", "-l"),
				},
			},
		},
		{
			name:  "pipe-quoted",
			lines: []string{`kubectl get pods --format=yaml | jq 'select(.conditions[]) | .status'`},
			expected: []Instruction{
				{
					Command: cmd.NewCmd("kubectl", "get", "pods", "--format=yaml"),
					Piped:   true,
				},
				{
					Command: cmd.NewCmd("jq", `select(.conditions[]) | .status`),
				},
			},
		},
		{
			name:  "nested-quotes",
			lines: []string{`gcloud logging read "resource.labels.project_id=\"foyle-dev\" resource.type=\"k8s_container\" resource.labels.location=\"us-west1\" resource.labels.cluster_name=\"dev\"" --project=foyle-dev`},
			expected: []Instruction{
				{
					Command: cmd.NewCmd("gcloud", "logging", "read", "resource.labels.project_id=\"foyle-dev\" resource.type=\"k8s_container\" resource.labels.location=\"us-west1\" resource.labels.cluster_name=\"dev\"", "--project=foyle-dev"),
				},
			},
		},
	}

	parser, err := NewBashishParser()

	if err != nil {
		t.Fatalf("NewBashishParser() returned error %v", err)
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			doc := strings.Join(c.lines, "\n")
			actual, err := parser.Parse(doc)
			if err != nil {
				t.Fatalf("unexpected parsing error %v", err)
			}
			if len(actual) != len(c.expected) {
				t.Errorf("Expected %v instructions got %v", len(c.expected), len(actual))
			}

			for i, eInstruction := range c.expected {
				if i >= len(actual) {
					break
				}

				aInstruction := actual[i]

				if aInstruction.Command.Name != eInstruction.Command.Name {
					t.Errorf("Expected command.Name to be %v got %v", eInstruction.Command.Name, aInstruction.Command.Name)
				}
				if d := cmp.Diff(eInstruction.Command.Args, aInstruction.Command.Args); d != "" {
					t.Fatalf("Unexpected args (-want +got): %v", d)
				}

				if aInstruction.Piped != eInstruction.Piped {
					t.Errorf("Expected Piped to be %v got %v", eInstruction.Piped, aInstruction.Piped)
				}
			}
		})
	}
}
