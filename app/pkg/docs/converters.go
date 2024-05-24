package docs

import (
	"strings"

	"github.com/jlewi/foyle/app/pkg/runme/converters"
	parserv1 "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/parser/v1"
	"github.com/stateful/runme/v3/pkg/document/identity"

	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/stateful/runme/v3/pkg/document/editor"
)

// BlockToMarkdown converts a block to markdown
func BlockToMarkdown(block *v1alpha1.Block) string {
	sb := strings.Builder{}
	writeBlockMarkdown(&sb, block)
	return sb.String()
}

func writeBlockMarkdown(sb *strings.Builder, block *v1alpha1.Block) {
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
}

// BlocksToMarkdown converts a sequence of blocks to markdown
func BlocksToMarkdown(blocks []*v1alpha1.Block) string {
	sb := strings.Builder{}

	for _, block := range blocks {
		writeBlockMarkdown(&sb, block)
	}

	return sb.String()
}

// DocToMarkdown converts a doc to markdown
func DocToMarkdown(doc *v1alpha1.Doc) string {
	return BlocksToMarkdown(doc.GetBlocks())
}

// MarkdownToBlocks converts a markdown string into a sequence of blocks.
// This function relies on RunMe's Markdown->Cells conversion; underneath the hood that uses goldmark to walk the AST.
// RunMe's deserialization function doesn't have any notion of output in markdown. However, in Foyle outputs
// are rendered to code blocks of language "output". So we need to do some post processing to convert the outputs
// into output items
func MarkdownToBlocks(mdText string) ([]*v1alpha1.Block, error) {
	// N.B. We don't need to add any identities
	resolver := identity.NewResolver(identity.UnspecifiedLifecycleIdentity)
	notebook, err := editor.Deserialize([]byte(mdText), resolver)

	blocks := make([]*v1alpha1.Block, 0, len(notebook.Cells))

	var lastCodeBlock *v1alpha1.Block
	for _, cell := range notebook.Cells {

		var tr *parserv1.TextRange

		if cell.TextRange != nil {
			tr = &parserv1.TextRange{
				Start: uint32(cell.TextRange.Start),
				End:   uint32(cell.TextRange.End),
			}
		}

		cellPb := &parserv1.Cell{
			Kind:       parserv1.CellKind(cell.Kind),
			Value:      cell.Value,
			LanguageId: cell.LanguageID,
			Metadata:   cell.Metadata,
			TextRange:  tr,
		}

		block, err := converters.CellToBlock(cellPb)
		if err != nil {
			return nil, err
		}

		// We need to handle the case where the block is an output code block.
		if block.Kind == v1alpha1.BlockKind_CODE {
			if block.Language == OUTPUTLANG {
				// This is an output block
				// We need to append the output to the last code block
				if lastCodeBlock != nil {
					if lastCodeBlock.Outputs == nil {
						lastCodeBlock.Outputs = make([]*v1alpha1.BlockOutput, 0, 1)
					}
					lastCodeBlock.Outputs = append(lastCodeBlock.Outputs, &v1alpha1.BlockOutput{
						Items: []*v1alpha1.BlockOutputItem{
							{
								TextData: block.Contents,
							},
						},
					})
					continue
				}

				// Since we don't have a code block to add the output to just treat it as a code block
			} else {
				// Update the lastCodeBlock
				lastCodeBlock = block
			}
		} else {
			// If we have a non-nil markup block then we zero out lastCodeBlock so that a subsequent output block
			// wouldn't be added to the last code block.
			if block.GetContents() != "" {
				lastCodeBlock = nil
			}
		}

		blocks = append(blocks, block)
	}

	return blocks, err
}
