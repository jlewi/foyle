package agent

import (
	"github.com/google/go-cmp/cmp"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Prompt(t *testing.T) {
	// Perhaps of this unittest is to verify that the template gets rendered the way we expect
	type testCase struct {
		args         promptArgs
		expectedFile string
	}

	cases := []testCase{
		{
			args: promptArgs{
				Document: "some document",
			},
			expectedFile: "no_examples.txt",
		},
		{
			args: promptArgs{
				Document: "blah blah",
				Examples: []Example{
					{
						Input:  "input1",
						Output: "output1",
					},
					{
						Input:  "input2",
						Output: "output2",
					},
				},
			},
			expectedFile: "examples.txt",
		},
	}

	updateExpected := (os.Getenv("UPDATE_EXPECTED") != "")

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	testDir := filepath.Join(cwd, "test_data")
	for _, c := range cases {
		t.Run(c.expectedFile, func(t *testing.T) {

			var sb strings.Builder
			if err := promptTemplate.Execute(&sb, c.args); err != nil {
				t.Fatalf("Failed to execute prompt template: %v", err)
			}

			actual := sb.String()
			expectedFile := filepath.Join(testDir, c.expectedFile)

			if updateExpected {
				t.Logf("Updating expected file %v", expectedFile)
				if err := os.WriteFile(expectedFile, []byte(actual), 0644); err != nil {
					t.Fatalf("Failed to write expected file: %v", err)
				}
			}

			expected, err := os.ReadFile(expectedFile)
			if err != nil {
				t.Fatalf("Failed to read expected file: %v", err)
			}

			if d := cmp.Diff(string(expected), actual); d != "" {
				t.Errorf("Unexpected diff:\n%s", d)
			}
		})
	}
}
