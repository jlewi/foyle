package docs

import (
	"context"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
)

// CreateQuery creates a query from a GenerateRequest
// It returns the blocks that should be used to query for similar documents
func CreateQuery(ctx context.Context, req *v1alpha1.GenerateRequest) ([]*v1alpha1.Block, error) {
	// Use a simple algorithm.
	// 1. Always select at least the current block
	// 2. Select additional blocks if they are markup blocks.
	startIndex := req.GetSelectedIndex() - 1

	for ; startIndex >= 0; startIndex-- {
		if req.GetDoc().GetBlocks()[startIndex].Kind != v1alpha1.BlockKind_MARKUP {
			break
		}
	}

	blocks := req.GetDoc().GetBlocks()[startIndex+1 : req.GetSelectedIndex()+1]
	return blocks, nil
}
