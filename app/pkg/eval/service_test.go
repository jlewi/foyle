package eval

import (
	"fmt"
	"testing"

	parserv1 "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/parser/v1"

	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
)

func Test_ToAssertRow(t *testing.T) {
	type testCase struct {
		evalResult *v1alpha1.EvalResult
		expected   *v1alpha1.AssertionRow
	}

	cases := []testCase{
		{
			evalResult: &v1alpha1.EvalResult{
				Example: &v1alpha1.EvalExample{
					Id: "1234",
					FullContext: &v1alpha1.FullContext{
						Notebook: &parserv1.Notebook{
							Cells: []*parserv1.Cell{
								{
									Kind:  parserv1.CellKind_CELL_KIND_MARKUP,
									Value: "Hello world",
								},
							},
						},
					},
				},
				ActualCells: []*parserv1.Cell{
					{
						Kind:  parserv1.CellKind_CELL_KIND_MARKUP,
						Value: "word",
					},
				},
				Assertions: []*v1alpha1.Assertion{
					{
						Name:   "AssertCodeAfterMarkdown",
						Result: v1alpha1.AssertResult_PASSED,
					},
				},
			},
			expected: &v1alpha1.AssertionRow{
				Id:                "1234",
				DocMd:             "Hello world\n",
				AnswerMd:          "word\n",
				CodeAfterMarkdown: v1alpha1.AssertResult_PASSED,
			},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("Case %d", i), func(t *testing.T) {
			actual, err := toAssertionRow(tc.evalResult)
			if err != nil {
				t.Errorf("Error converting %v; %v", tc.evalResult, err)
			}
			if actual.Id != tc.expected.Id {
				t.Errorf("Unexpected Id; got %v, want %v", actual.Id, tc.expected.Id)
			}
			if actual.DocMd != tc.expected.DocMd {
				t.Errorf("Unexpected DocMd; got %v, want %v", actual.DocMd, tc.expected.DocMd)
			}
			if actual.AnswerMd != tc.expected.AnswerMd {
				t.Errorf("Unexpected AnswerMd; got %v, want %v", actual.AnswerMd, tc.expected.AnswerMd)
			}
			if actual.CodeAfterMarkdown != tc.expected.CodeAfterMarkdown {
				t.Errorf("Unexpected CodeAfterMarkdown; got %v, want %v", actual.CodeAfterMarkdown, tc.expected.CodeAfterMarkdown)
			}
		})
	}
}
