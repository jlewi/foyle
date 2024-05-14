package runme

import (
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/pkg/errors"
)
import "github.com/jlewi/foyle/runme/gen/proto/go/foyle/runme"

// NotebookToDoc converts a runme Notebook to a foyle Doc
func NotebookToDoc(nb *runme.Notebook) (*v1alpha1.Doc, error) {
	if nb == nil {
		return nil, errors.New("Notebook is nil")
	}

	doc := &v1alpha1.Doc{
		Blocks: make([]*v1alpha1.Block, 0, len(nb.Cells)),
	}

	for _, cell := range nb.Cells {
		block, err := CellToBlock(cell)
		if err != nil {
			return nil, err
		}
		doc.Blocks = append(doc.Blocks, block)
	}

	return doc, nil
}

// CellToBlock converts a runme Cell to a foyle Block
//
// N.B. cell metadata is currently ignored.
func CellToBlock(cell *runme.Cell) (*v1alpha1.Block, error) {
	if cell == nil {
		return nil, errors.New("Cell is nil")
	}

	blockOutputs := make([]*v1alpha1.BlockOutput, 0, len(cell.Outputs))

	for _, output := range cell.Outputs {
		bOutput, err := CellOutputToBlockOutput(output)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to convert CellOutput to BlockOutput")
		}
		blockOutputs = append(blockOutputs, bOutput)
	}
	blockKind := CellKindToBlockKind(cell.Kind)
	return &v1alpha1.Block{
		Language: cell.LanguageId,
		Contents: cell.Value,
		Kind:     blockKind,
		Outputs:  blockOutputs,
	}, nil
}

func CellKindToBlockKind(kind runme.CellKind) v1alpha1.BlockKind {
	switch kind {
	case runme.CellKind_CELL_KIND_CODE:
		return v1alpha1.BlockKind_CODE
	case runme.CellKind_CELL_KIND_MARKUP:
		return v1alpha1.BlockKind_MARKUP
	default:
		return v1alpha1.BlockKind_UNKNOWN_BLOCK_KIND
	}
}

func CellOutputToBlockOutput(output *runme.CellOutput) (*v1alpha1.BlockOutput, error) {
	if output == nil {
		return nil, errors.New("CellOutput is nil")
	}

	boutput := &v1alpha1.BlockOutput{
		Items: make([]*v1alpha1.BlockOutputItem, 0, len(output.Items)),
	}

	for _, oi := range output.Items {
		boi := &v1alpha1.BlockOutputItem{
			Mime:     oi.Mime,
			TextData: string(oi.Data),
		}
		boutput.Items = append(boutput.Items, boi)
	}

	return boutput, nil
}
