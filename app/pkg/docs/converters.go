package docs

import (
	"strings"

	markdownfmt "github.com/Kunde21/markdownfmt/v3/markdown"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// BlockToMarkdown converts a block to markdown
func BlockToMarkdown(block *v1alpha1.Block) string {
	sb := strings.Builder{}

	switch block.GetKind() {
	case v1alpha1.BlockKind_CODE:
		// Code just gets written as a code block
		sb.WriteString("```" + BASHLANG + "\n")
		sb.WriteString(block.GetContents())
		sb.WriteString("\n```\n")
	default:
		// Otherwise assume its a markdown block
		sb.WriteString(block.GetContents() + "\n")
	}

	// Handle the outputs
	for _, output := range block.GetOutputs() {
		for _, oi := range output.Items {
			sb.WriteString("```" + OUTPUTLANG + "\n")
			sb.WriteString(oi.GetTextData())
			sb.WriteString("\n```\n")
		}
	}

	return sb.String()
}

// MarkdownToBlocks converts a markdown string into a sequence of blocks.
// This function relies on the goldmark library to parse the markdown into an AST.
func MarkdownToBlocks(mdText string) ([]*v1alpha1.Block, error) {
	gm := goldmark.New()
	source := []byte(mdText)
	reader := text.NewReader(source)
	root := gm.Parser().Parse(reader)

	renderer := markdownfmt.NewRenderer()

	blocks := make([]*v1alpha1.Block, 0, 20)

	err := ast.Walk(root, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			// Do nothing on leaving the block; just continue the walk
			return ast.WalkContinue, nil
		}

		if node.Kind() == ast.KindDocument {
			// Ignore the document node
			return ast.WalkContinue, nil
		}

		if node.Kind() != ast.KindFencedCodeBlock {
			// Since we aren't in a code block render the node and its children to markdown
			// so we can add them as a block
			var sb strings.Builder
			if err := renderer.Render(&sb, source, node); err != nil {
				return ast.WalkStop, err
			}
			newBlock := &v1alpha1.Block{
				Kind:     v1alpha1.BlockKind_MARKUP,
				Contents: sb.String(),
			}
			blocks = append(blocks, newBlock)
			// Skip the children because we've already rendered the children to markdown so there's no need
			// to visit the children nodes
			return ast.WalkSkipChildren, nil

		}

		// Since we encountered a fenced code block we need to extract the code block
		fenced := node.(*ast.FencedCodeBlock)
		lang := string(fenced.Language(source))
		textData := getBlockText(fenced, source)

		lastBlock := len(blocks) - 1
		lastWasCode := false
		if lastBlock >= 0 && blocks[lastBlock].Kind == v1alpha1.BlockKind_CODE {
			lastWasCode = true
		}

		if lang == OUTPUTLANG && lastWasCode {
			// Since its an output block and the last block was a code block we should append the output to the last block
			if blocks[lastBlock].Outputs == nil {
				blocks[lastBlock].Outputs = make([]*v1alpha1.BlockOutput, 0, 1)
			}
			blocks[lastBlock].Outputs = append(blocks[lastBlock].Outputs, &v1alpha1.BlockOutput{
				Items: []*v1alpha1.BlockOutputItem{
					{
						TextData: textData,
					},
				},
			})
		} else {
			block := &v1alpha1.Block{
				Kind:     v1alpha1.BlockKind_CODE,
				Contents: textData,
				Language: lang,
			}
			blocks = append(blocks, block)
		}

		// We can skip walking the children of the code block since we've already ingested the code block
		return ast.WalkSkipChildren, nil
	})

	// The way we walk the AST above we potentially end up segmenting continuous markdown without code blocks
	// into more than one block. So we merge these blocks.
	final := make([]*v1alpha1.Block, 0, len(blocks))
	i := 0
	for _, block := range blocks {
		lastBlock := i - 1
		addToLastBlock := false
		if lastBlock >= 0 && block.Kind == v1alpha1.BlockKind_MARKUP && final[lastBlock].Kind == v1alpha1.BlockKind_MARKUP {
			addToLastBlock = true
		}

		if addToLastBlock {
			final[lastBlock].Contents += block.Contents
		} else {
			final = append(final, block)
			i++
		}
	}

	return final, err
}

func getBlockText(fenced *ast.FencedCodeBlock, source []byte) string {
	var sb strings.Builder
	for i := 0; i < fenced.Lines().Len(); i++ {
		// Get the i'th line
		line := fenced.Lines().At(i)
		sb.WriteString(string(line.Value(source)))
	}
	return sb.String()
}
