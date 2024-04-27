package docs

import (
	"strings"

	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
)

// Tailer is a helper for building a markdown representation of the tail end of a document.
// It is intended to be stateful and used to iteratively find a suffix of a document that fits within a certain length
// (i.e. the context length of the model).
type Tailer struct {
	// mdBlocks keeps track of the markdown blocks
	mdBlocks []string

	// firstBlock is the index of the first block to include in the prompt
	firstBlock int
}

func NewTailer(blocks []*v1alpha1.Block, maxCharLen int) *Tailer {
	mdBlocks := make([]string, len(blocks))

	length := 0
	firstBlock := len(blocks) - 1
	for ; firstBlock >= 0; firstBlock-- {
		block := blocks[firstBlock]
		md := BlockToMarkdown(block)
		if length+len(md) > maxCharLen {
			break
		}
		mdBlocks[firstBlock] = md
	}

	return &Tailer{
		mdBlocks: mdBlocks,
	}
}

// Text returns the text of the doc.
func (p *Tailer) Text() string {
	var sb strings.Builder
	for i := p.firstBlock; i < len(p.mdBlocks); i++ {
		// N.B. we need to keep this in sync BlocksToMarkdown w.r.t. inserting new whitespace. Otherwise
		// we could potentially introduce drift in our data.
		sb.WriteString(p.mdBlocks[i])
	}
	return sb.String()
}

// Shorten shortens the doc that will be generated on the next call to Text.
// Return false if the doc can't be shortened any further.
func (p *Tailer) Shorten() bool {
	if p.firstBlock+1 >= len(p.mdBlocks) {
		return false
	}

	p.firstBlock += 1
	return true
}
