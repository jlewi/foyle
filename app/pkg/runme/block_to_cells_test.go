package runme

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	parserv1 "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/parser/v1"
	"testing"
)

func Test_BlocksToCells(t *testing.T) {
	for _, c := range cases {
		t.Run(fmt.Sprintf("Case %s", c.name), func(t *testing.T) {
			actual, err := BlocksToCells(c.Doc.Blocks)
			if err != nil {
				t.Errorf("Case %v: Error %v", c.name, err)
				return
			}

			if diff := cmp.Diff(c.Notebook.Cells, actual, cmpopts.IgnoreUnexported(parserv1.Cell{}, parserv1.CellOutput{}, parserv1.CellOutputItem{})); diff != "" {
				t.Errorf("Unexpected Diff:\n%v", diff)
			}
		})
	}
}
