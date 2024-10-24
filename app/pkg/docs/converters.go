package docs

import (
	"strings"

	"github.com/jlewi/foyle/app/pkg/runme/converters"
	parserv1 "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/parser/v1"
	"github.com/stateful/runme/v3/pkg/document/identity"

	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/stateful/runme/v3/pkg/document/editor"
)

const (
	truncationMessage = "<...stdout was truncated...>"
)

// BlockToMarkdown converts a block to markdown
// maxOutputLength is the maximum length of the output to include. A value of <= 0 means no truncation.
func BlockToMarkdown(block *v1alpha1.Block, maxOutputLength int) string {
	sb := strings.Builder{}
	writeBlockMarkdown(&sb, block, maxOutputLength)
	return sb.String()
}

func writeBlockMarkdown(sb *strings.Builder, block *v1alpha1.Block, maxOutputLength int) {
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
			if oi.GetMime() == StatefulRunmeOutputItemsMimeType || oi.GetMime() == StatefulRunmeTerminalMimeType {
				// See: https://github.com/jlewi/foyle/issues/286. This output item contains a JSON dictionary
				// with a bunch of meta information that seems specific to Runme/stateful and not necessarily
				// relevant as context for AI so we filter it out. The output item we are interested in should
				// have a mime type of application/vnd.code.notebook.stdout and contain the stdout of the executed
				// code.
				//
				// We use an exclude list for now because Runme is adding additional mime types as it adds custom
				// renderers. https://github.com/stateful/vscode-runme/blob/3e36b16e3c41ad0fa38f0197f1713135e5edb27b/src/constants.ts#L6
				// So for now we want to error on including useless data rather than silently dropping useful data.
				// In the future we may want to revisit that.
				//
				// N.B. On the other hand our code for truncating long outputs is based on searching for the mime
				// type application/vnd.code.notebook.stdout so we need to be careful to ensure that we don't
				continue
			}

			sb.WriteString("```" + OUTPUTLANG + "\n")
			textData := oi.GetTextData()
			if 0 < maxOutputLength && len(textData) > maxOutputLength {
				textData = textData[:maxOutputLength]
				sb.WriteString(textData)
				// Don't write a newline before writing truncation because that is more likely to lead to confusion
				// because people might not realize the line was truncated.
				// Emit a message indicating that the output was truncated
				// This is intended for the LLM so it knows that it is working with a truncated output.
				sb.WriteString(truncationMessage)
			} else {
				sb.WriteString(textData)
			}

			sb.WriteString("\n```\n")
		}
	}
}

// BlocksToMarkdown converts a sequence of blocks to markdown
func BlocksToMarkdown(blocks []*v1alpha1.Block) string {
	sb := strings.Builder{}

	for _, block := range blocks {
		writeBlockMarkdown(&sb, block, -1)
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
