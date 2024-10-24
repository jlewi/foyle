package docs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jlewi/foyle/app/pkg/testutil"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
)

func Test_BlockToMarkdown(t *testing.T) {
	type testCase struct {
		name         string
		block        *v1alpha1.Block
		maxOutputLen int
		expected     string
	}

	testCases := []testCase{
		{
			name: "markup",
			block: &v1alpha1.Block{
				Kind:     v1alpha1.BlockKind_MARKUP,
				Contents: "This is a test",
			},
			expected: "This is a test\n",
		},
		{
			name: "code",
			block: &v1alpha1.Block{
				Kind:     v1alpha1.BlockKind_CODE,
				Contents: "echo \"something something\"",
				Outputs: []*v1alpha1.BlockOutput{
					{
						Items: []*v1alpha1.BlockOutputItem{
							{
								TextData: "something something",
							},
						},
					},
				},
			},
			expected: "```bash\necho \"something something\"\n```\n```output\nsomething something\n```\n",
		},
		{
			name: "filter-by-mime-type",
			block: &v1alpha1.Block{
				Kind:     v1alpha1.BlockKind_CODE,
				Contents: "echo \"something something\"",
				Outputs: []*v1alpha1.BlockOutput{
					{
						Items: []*v1alpha1.BlockOutputItem{
							{
								TextData: "Should be excluded",
								Mime:     StatefulRunmeOutputItemsMimeType,
							},
							{
								TextData: "Terminal be excluded",
								Mime:     StatefulRunmeTerminalMimeType,
							},
							{
								TextData: "Should be included",
								Mime:     "application/vnd.code.notebook.stdout",
							},
						},
					},
				},
			},
			expected: "```bash\necho \"something something\"\n```\n```output\nShould be included\n```\n",
		},
		{
			name: "truncate-output",
			block: &v1alpha1.Block{
				Kind:     v1alpha1.BlockKind_CODE,
				Contents: "echo \"something something\"",
				Outputs: []*v1alpha1.BlockOutput{
					{
						Items: []*v1alpha1.BlockOutputItem{
							{
								TextData: "some really long output",
							},
						},
					},
				},
			},
			maxOutputLen: 5,
			expected:     "```bash\necho \"something something\"\n```\n```output\nsome <...stdout was truncated...>\n```\n",
		},
	}
	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			actual := BlockToMarkdown(c.block, c.maxOutputLen)
			if d := cmp.Diff(c.expected, actual); d != "" {
				t.Errorf("Unexpected diff:\n%s", d)
			}
		})
	}
}

func Test_DockToMarkdown(t *testing.T) {
	type testCase struct {
		name     string
		doc      *v1alpha1.Doc
		expected string
	}

	// Most test cases are covered by BlockToMarkdown. The primary purpose of this test is to ensure proper
	// spacing between blocks.
	testCases := []testCase{
		{
			name: "markup",
			doc: &v1alpha1.Doc{
				Blocks: []*v1alpha1.Block{
					{
						Kind:     v1alpha1.BlockKind_MARKUP,
						Contents: "block 1",
					},
					{
						Kind:     v1alpha1.BlockKind_MARKUP,
						Contents: "block 2",
					},
				},
			},
			expected: "block 1\nblock 2\n",
		},
	}
	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			actual := DocToMarkdown(c.doc)
			if d := cmp.Diff(c.expected, actual); d != "" {
				t.Errorf("Unexpected diff:\n%s", d)
			}
		})
	}
}
func Test_MarkdownToBlocks(t *testing.T) {
	type testCase struct {
		name     string
		inFile   string
		expected []*v1alpha1.Block
	}

	cases := []testCase{
		{
			name:   "simple",
			inFile: "testdoc.md",
			expected: []*v1alpha1.Block{
				{
					Kind:     v1alpha1.BlockKind_MARKUP,
					Contents: "# Section 1",
					Outputs:  []*v1alpha1.BlockOutput{},
				},
				{
					Kind:     v1alpha1.BlockKind_MARKUP,
					Contents: "This is section 1",
					Outputs:  []*v1alpha1.BlockOutput{},
				},
				{
					Kind:     v1alpha1.BlockKind_CODE,
					Language: "go",
					Contents: "package main\n\nfunc main() {\n...\n}",
					Outputs:  []*v1alpha1.BlockOutput{},
				},
				{
					Kind:     v1alpha1.BlockKind_MARKUP,
					Contents: "Breaking text",
					Outputs:  []*v1alpha1.BlockOutput{},
				},
				{
					Kind:     v1alpha1.BlockKind_CODE,
					Language: "bash",
					Contents: "echo \"Hello, World!\"",
					Outputs: []*v1alpha1.BlockOutput{
						{
							Items: []*v1alpha1.BlockOutputItem{
								{
									TextData: "hello, world!",
								}},
						},
					},
				},
				{
					Kind:     v1alpha1.BlockKind_MARKUP,
					Contents: "## Subsection",
					Outputs:  []*v1alpha1.BlockOutput{},
				},
			},
		},
		{
			name:   "list-nested",
			inFile: "list.md",
			expected: []*v1alpha1.Block{
				{
					Kind:     v1alpha1.BlockKind_MARKUP,
					Contents: "Test code blocks nested in a list",
					Outputs:  []*v1alpha1.BlockOutput{},
				},
				{
					Kind:     v1alpha1.BlockKind_MARKUP,
					Contents: "1. First command",
					Outputs:  []*v1alpha1.BlockOutput{},
				},
				{
					Kind:     v1alpha1.BlockKind_CODE,
					Language: "bash",
					Contents: "echo 1",
					Outputs:  []*v1alpha1.BlockOutput{},
				},
				{
					Kind:     v1alpha1.BlockKind_MARKUP,
					Contents: "2. Second command",
					Outputs:  []*v1alpha1.BlockOutput{},
				},
				{
					Kind:     v1alpha1.BlockKind_CODE,
					Language: "bash",
					Contents: "echo 2",
					Outputs:  []*v1alpha1.BlockOutput{},
				},
			},
		},
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fPath := filepath.Join(cwd, "test_data", c.inFile)
			raw, err := os.ReadFile(fPath)
			if err != nil {
				t.Fatalf("Failed to read raw file: %v", err)
			}
			actual, err := MarkdownToBlocks(string(raw))
			if err != nil {
				t.Fatalf("MarkdownToBlocks(%v) returned error %v", c.inFile, err)
			}
			if len(actual) != len(c.expected) {
				t.Errorf("Expected %v blocks got %v", len(c.expected), len(actual))
			}

			for i, eBlock := range c.expected {
				if i >= len(actual) {
					break
				}

				aBlock := actual[i]

				if d := cmp.Diff(eBlock, aBlock, testutil.BlockComparer); d != "" {
					t.Errorf("Unexpected diff block %d:\n%s", i, d)
				}
			}
		})
	}
}
