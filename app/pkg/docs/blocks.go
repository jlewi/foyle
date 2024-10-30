package docs

import (
	"context"

	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
)

// CreateQuery creates a query from a GenerateRequest
// It returns the blocks that should be used to query for similar documents
func CreateQuery(ctx context.Context, req *v1alpha1.GenerateRequest) ([]*v1alpha1.Block, error) {
	// PreprocessDoc trims the document to only include blocks up to the selected index
	blocks := PreprocessDoc(req)
	// Use a simple algorithm.
	// 1. Always select at least the current block
	// 2. Select additional blocks if they are markup blocks.
	startIndex := len(blocks) - 2

	for ; startIndex >= 0; startIndex-- {
		if blocks[startIndex].Kind != v1alpha1.BlockKind_MARKUP {
			break
		}
	}

	// startIndex should be pointing at a block we don't want to include so we add 1 to it to get the first block
	// to include
	blocks = blocks[startIndex+1:]
	return blocks, nil
}
