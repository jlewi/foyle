package runme

import (
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/pkg/errors"
	parserv1 "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/parser/v1"
)

func BlocksToCells(blocks []*v1alpha1.Block) ([]*parserv1.Cell, error) {
	cells := make([]*parserv1.Cell, 0, len(blocks))

	for _, block := range blocks {
		cell, err := BlockToCell(block)
		if err != nil {
			return nil, err
		}
		cells = append(cells, cell)
	}
	return cells, nil
}

// BlockToCell converts a foyle Block to a RunMe Cell
func BlockToCell(block *v1alpha1.Block) (*parserv1.Cell, error) {
	if block == nil {
		return nil, errors.New("block is nil")
	}

	cellOutputs := make([]*parserv1.CellOutput, 0, len(block.Outputs))

	for _, output := range block.Outputs {
		cOutput, err := BlockOutputToCellOutput(output)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to convert BlockOutput to CellOutput")
		}
		cellOutputs = append(cellOutputs, cOutput)
	}
	cellKind := BlockKindToCellKind(block.Kind)
	return &parserv1.Cell{
		LanguageId: block.Language,
		Value:      block.Contents,
		Kind:       cellKind,
		Outputs:    cellOutputs,
	}, nil
}

func BlockKindToCellKind(kind v1alpha1.BlockKind) parserv1.CellKind {
	switch kind {
	case v1alpha1.BlockKind_CODE:
		return parserv1.CellKind_CELL_KIND_CODE
	case v1alpha1.BlockKind_MARKUP:
		return parserv1.CellKind_CELL_KIND_MARKUP
	default:
		return parserv1.CellKind_CELL_KIND_UNSPECIFIED
	}
}
func BlockOutputToCellOutput(output *v1alpha1.BlockOutput) (*parserv1.CellOutput, error) {
	if output == nil {
		return nil, errors.New("BlockOutput is nil")
	}

	coutput := &parserv1.CellOutput{
		Items: make([]*parserv1.CellOutputItem, 0, len(output.Items)),
	}

	for _, oi := range output.Items {
		boi := &parserv1.CellOutputItem{
			Mime: oi.Mime,
			Data: []byte(oi.TextData),
		}
		coutput.Items = append(coutput.Items, boi)
	}

	return coutput, nil
}
