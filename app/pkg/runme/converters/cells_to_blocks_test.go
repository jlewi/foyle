package converters

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jlewi/foyle/app/pkg/testutil"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	parserv1 "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/parser/v1"
)

type testCase struct {
	name     string
	Notebook *parserv1.Notebook
	Doc      *v1alpha1.Doc
}

var (
	cases = []testCase{
		{
			name: "Simple",
			Notebook: &parserv1.Notebook{
				Cells: []*parserv1.Cell{
					{
						Metadata: map[string]string{
							"id":          "1234",
							"interactive": "false",
						},
						Kind:       parserv1.CellKind_CELL_KIND_CODE,
						LanguageId: "python",
						Value:      "print('Hello World')",
						Outputs: []*parserv1.CellOutput{
							{
								Items: []*parserv1.CellOutputItem{
									{
										Data: []byte("Hello World\n"),
										Mime: "text/plain",
									},
								},
							},
						},
					},
				},
			},
			Doc: &v1alpha1.Doc{
				Blocks: []*v1alpha1.Block{
					{
						Id:       "1234",
						Language: "python",
						Contents: "print('Hello World')",
						Metadata: map[string]string{
							"id":          "1234",
							"interactive": "false",
						},
						Kind: v1alpha1.BlockKind_CODE,
						Outputs: []*v1alpha1.BlockOutput{
							{
								Items: []*v1alpha1.BlockOutputItem{
									{
										TextData: "Hello World\n",
										Mime:     "text/plain",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			// This test case we don't set interactive explicitly.
			// It verifies its not getting added
			name: "no-interactive",
			Notebook: &parserv1.Notebook{
				Cells: []*parserv1.Cell{
					{
						Metadata: map[string]string{
							"id": "1234",
						},
						Kind:       parserv1.CellKind_CELL_KIND_CODE,
						LanguageId: "python",
						Value:      "print('Hello World')",
						Outputs: []*parserv1.CellOutput{
							{
								Items: []*parserv1.CellOutputItem{
									{
										Data: []byte("Hello World\n"),
										Mime: "text/plain",
									},
								},
							},
						},
					},
				},
			},
			Doc: &v1alpha1.Doc{
				Blocks: []*v1alpha1.Block{
					{
						Id:       "1234",
						Language: "python",
						Contents: "print('Hello World')",
						Metadata: map[string]string{
							"id": "1234",
						},
						Kind: v1alpha1.BlockKind_CODE,
						Outputs: []*v1alpha1.BlockOutput{
							{
								Items: []*v1alpha1.BlockOutputItem{
									{
										TextData: "Hello World\n",
										Mime:     "text/plain",
									},
								},
							},
						},
					},
				},
			},
		},
	}
)

func Test_NotebookToDoc(t *testing.T) {
	for i, c := range cases {
		t.Run(fmt.Sprintf("Case %d", i), func(t *testing.T) {
			actual, err := NotebookToDoc(c.Notebook)
			if err != nil {
				t.Errorf("Case %v: Error %v", i, err)
				return
			}

			if diff := cmp.Diff(c.Doc, actual, testutil.DocComparer); diff != "" {
				t.Errorf("Unexpected Diff:\n%v", diff)
			}
		})
	}
}
