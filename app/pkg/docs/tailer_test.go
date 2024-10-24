package docs

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
)

func Test_Tailer(t *testing.T) {
	type testCase struct {
		name     string
		Doc      *v1alpha1.Doc
		MaxChars int
		Expected string
	}

	cases := []testCase{
		{
			name: "cell-longer-than-max-chars",
			Doc: &v1alpha1.Doc{
				Blocks: []*v1alpha1.Block{
					{
						Kind:     v1alpha1.BlockKind_MARKUP,
						Contents: "Cell1",
					},
					{
						Kind:     v1alpha1.BlockKind_MARKUP,
						Contents: "Cell2\nCell3",
					},
				},
			},
			MaxChars: 5,
			Expected: "Cell3\n",
		},
		{
			name: "multiple-cells",
			Doc: &v1alpha1.Doc{
				Blocks: []*v1alpha1.Block{
					{
						Kind:     v1alpha1.BlockKind_MARKUP,
						Contents: "Cell1",
					},
					{
						Kind:     v1alpha1.BlockKind_MARKUP,
						Contents: "Cell2",
					},
					{
						Kind:     v1alpha1.BlockKind_MARKUP,
						Contents: "Cell3",
					},
				},
			},
			MaxChars: 12,
			Expected: "Cell2\nCell3\n",
		},
		{
			name: "truncate-outputs",
			Doc: &v1alpha1.Doc{
				Blocks: []*v1alpha1.Block{
					{
						Kind:     v1alpha1.BlockKind_CODE,
						Contents: "Cell1",
						Outputs: []*v1alpha1.BlockOutput{
							{
								Items: []*v1alpha1.BlockOutputItem{
									{
										TextData: "Output1\nOutput2\nOutput3",
										Mime:     VSCodeNotebookStdOutMimeType,
									},
								},
							},
						},
					},
				},
			},
			MaxChars: 12,
			Expected: "```bash\nCell1\n```\n```output\nOutput1<...stdout was truncated...>\n```\n",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tailer := NewTailer(context.Background(), c.Doc.Blocks, c.MaxChars)
			actual := tailer.Text()
			if d := cmp.Diff(c.Expected, actual); d != "" {
				t.Fatalf("Unexpected diff:\n%v", d)
			}
		})
	}
}

func Test_tailLines(t *testing.T) {
	type testCase struct {
		name     string
		Contents string
		MaxChars int
		Expected string
	}

	cases := []testCase{
		{
			name:     "last-line-exceeds-max-chars",
			Contents: "line1\nline2",
			MaxChars: 2,
			Expected: "line2",
		},
		{
			name:     "all-lines",
			Contents: "line1\nline2",
			MaxChars: 30,
			Expected: "line1\nline2",
		},
		{
			name:     "some-lines",
			Contents: "line1\nline2\nline3",
			MaxChars: 10,
			Expected: "line2\nline3",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if tailLines(c.Contents, c.MaxChars) != c.Expected {
				t.Fatalf("Expected text to be %s but got %s", c.Expected, tailLines(c.Contents, c.MaxChars))
			}
		})
	}
}
