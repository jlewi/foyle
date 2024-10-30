package docs

import (
	"strconv"

	"github.com/jlewi/foyle/app/pkg/runme/converters"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
)

// PreprocessDoc does some preprocessing of the doc.
// It removes ghost cells and all cells after the selected index. Therefore the position of the cells in the returned
// array may not match original positions.
func PreprocessDoc(req *v1alpha1.GenerateRequest) []*v1alpha1.Block {
	// We want to remove all cells after the selected cell because our prompt doesn't know how to take them into account.
	// We also need to remove any Ghost cells
	cells := make([]*v1alpha1.Block, 0, req.GetSelectedIndex()+1)

	for i := int32(0); i < req.GetSelectedIndex()+1; i++ {
		c := req.Doc.GetBlocks()[i]
		metadata := c.Metadata
		if metadata == nil {
			metadata = make(map[string]string)
		}

		if v, ok := metadata[converters.GhostKeyField]; ok {
			// ParseBool returns error if the value is not a valid boolean value. In this case we just consider
			// it to not be a ghost cell
			if isGhost, err := strconv.ParseBool(v); err == nil && isGhost {
				// Ignore ghost cells
				continue
			}
		}

		cells = append(cells, c)
	}

	return cells
}
