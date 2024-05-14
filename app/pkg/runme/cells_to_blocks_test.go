package runme

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/jlewi/foyle/app/pkg/testutil"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	parserv1 "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/parser/v1"
	"testing"
)

func Test_NotebookToDoc(t *testing.T) {
	type testCase struct {
		Input    *parserv1.Notebook
		Expected *v1alpha1.Doc
	}

	cases := []testCase{
		{
			Input: &parserv1.Notebook{
				Cells: []*parserv1.Cell{
					{
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
			Expected: &v1alpha1.Doc{
				Blocks: []*v1alpha1.Block{
					{
						Language: "python",
						Contents: "print('Hello World')",
						Kind:     v1alpha1.BlockKind_CODE,
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

	for i, c := range cases {
		t.Run(fmt.Sprintf("Case %d", i), func(t *testing.T) {
			actual, err := NotebookToDoc(c.Input)
			if err != nil {
				t.Errorf("Case %v: Error %v", i, err)
				return
			}

			if diff := cmp.Diff(c.Expected, actual, testutil.DocComparer); diff != "" {
				t.Errorf("Unexpected Diff:\n%v", diff)
			}
		})
	}
}
