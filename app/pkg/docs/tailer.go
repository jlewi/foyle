package docs

import (
	"context"
	"strings"

	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/app/pkg/runme/ulid"

	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
)

const (
	numCodeBlockChars = len("```" + OUTPUTLANG + "\n" + "```\n")
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

func NewTailer(ctx context.Context, blocks []*v1alpha1.Block, maxCharLen int) *Tailer {
	log := logs.FromContext(ctx)
	mdBlocks := make([]string, len(blocks))

	firstBlock := len(blocks) - 1

	assertion := &v1alpha1.Assertion{
		Name:   v1alpha1.Assertion_AT_LEAST_ONE_FULL_INPUT_CELL,
		Result: v1alpha1.AssertResult_PASSED,
		Detail: "",
		Id:     ulid.GenerateID(),
	}

	numBlocks := 0
	for ; firstBlock >= 0 && maxCharLen > 0; firstBlock-- {
		block := blocks[firstBlock]
		numBlocks += 1
		md := BlockToMarkdown(block, maxCharLen)
		maxCharLen = maxCharLen - len(md)
		if maxCharLen <= 0 && numBlocks == 1 {
			// Since this is the first block and its truncated we fail the assertion.
			assertion.Result = v1alpha1.AssertResult_FAILED
		}

		mdBlocks[firstBlock] = md
	}

	log.Info(logs.Level1Assertion, "assertion", assertion)
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

// tailLines should always return a non-empty string if the input is non-empty.
// s is longer than maxLen then it will attempt to return the last n lines of s.
func tailLines(s string, maxLen int) string {
	lines := strings.Split(s, "\n")

	startIndex := len(lines) - 1

	length := len(lines[len(lines)-1])

	for ; startIndex >= 1; startIndex-- {
		nextIndex := startIndex - 1
		if len(lines[nextIndex])+length > maxLen {
			break
		}

		length += len(lines[nextIndex])
	}

	if startIndex < 0 {
		startIndex = 0
	}

	return strings.Join(lines[startIndex:], "\n")
}
