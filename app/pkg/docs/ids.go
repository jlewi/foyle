package docs

import (
	"github.com/google/uuid"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
)

// SetBlockIds sets the block ids for an blocks that don't have an id set.
// If a block already has an id set it will be left unchanged.
// The function returns the ids of the blocks.
func SetBlockIds(blocks []*v1alpha1.Block) ([]string, error) {
	blockIds := make([]string, 0, len(blocks))

	for i := range blocks {
		if blocks[i].Id == "" {
			newUid, err := uuid.NewRandom()
			if err != nil {
				return nil, err
			}
			blocks[i].Id = newUid.String()
		}
		blockIds = append(blockIds, blocks[i].Id)
	}

	return blockIds, nil
}
