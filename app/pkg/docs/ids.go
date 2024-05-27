package docs

import (
	"github.com/jlewi/foyle/app/pkg/runme/ulid"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
)

// SetBlockIds sets the block ids for an blocks that don't have an id set.
// If a block already has an id set it will be left unchanged.
// The function returns the ids of the blocks.
func SetBlockIds(blocks []*v1alpha1.Block) ([]string, error) {
	blockIds := make([]string, 0, len(blocks))

	for i := range blocks {
		if blocks[i].Id == "" {
			// We use ULIDs for block ids because that's what RunMe expects
			blocks[i].Id = ulid.GenerateID()
		}
		blockIds = append(blockIds, blocks[i].Id)
	}

	return blockIds, nil
}
