package docs

import (
	"github.com/jlewi/foyle/app/pkg/runme/converters"
	parserv1 "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/parser/v1"
	"github.com/stateful/runme/v3/pkg/document/identity"
	"strings"

	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/stateful/runme/v3/pkg/document/editor"
	"github.com/yuin/goldmark/ast"
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
// This function relies on the goldmark library to parse the markdown into an AST.
func MarkdownToBlocks(mdText string) ([]*v1alpha1.Block, error) {
	// N.B. We don't need to add any identities
	resolver := identity.NewResolver(identity.UnspecifiedLifecycleIdentity)
	notebook, err := editor.Deserialize([]byte(mdText), resolver)

	blocks := make([]*v1alpha1.Block, 0, len(notebook.Cells))

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
		blocks = append(blocks, block)
	}

	return blocks, err
	//gm := goldmark.New()
	//source := []byte(mdText)
	//reader := text.NewReader(source)
	//root := gm.Parser().Parse(reader)
	//
	//renderer := markdownfmt.NewRenderer()
	//
	//blocks := make([]*v1alpha1.Block, 0, 20)
	//
	//ast.DumpHelper(root, source, 0, nil, nil)
	//err := ast.Walk(root, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
	//	if !entering {
	//		// Do nothing on leaving the block; just continue the walk
	//		return ast.WalkContinue, nil
	//	}
	//
	//	if node.Kind() == ast.KindDocument {
	//		// Ignore the document node
	//		return ast.WalkContinue, nil
	//	}
	//
	//	if node.Kind() != ast.KindFencedCodeBlock {
	//		// Since we aren't in a code block render the node and its children to markdown
	//		// so we can add them as a block
	//		var sb strings.Builder
	//		if err := renderer.Render(&sb, source, node); err != nil {
	//			return ast.WalkStop, err
	//		}
	//
	//		newBlock := &v1alpha1.Block{
	//			Kind:     v1alpha1.BlockKind_MARKUP,
	//			Contents: sb.String(),
	//		}
	//		blocks = append(blocks, newBlock)
	//		// Skip the children because we've already rendered the children to markdown so there's no need
	//		// to visit the children nodes
	//		return ast.WalkSkipChildren, nil
	//
	//	}
	//
	//	// Since we encountered a fenced code block we need to extract the code block
	//	fenced := node.(*ast.FencedCodeBlock)
	//	lang := string(fenced.Language(source))
	//	textData := getBlockText(fenced, source)
	//
	//	lastBlock := len(blocks) - 1
	//	lastWasCode := false
	//	if lastBlock >= 0 && blocks[lastBlock].Kind == v1alpha1.BlockKind_CODE {
	//		lastWasCode = true
	//	}
	//
	//	if lang == OUTPUTLANG && lastWasCode {
	//		// Since its an output block and the last block was a code block we should append the output to the last block
	//		if blocks[lastBlock].Outputs == nil {
	//			blocks[lastBlock].Outputs = make([]*v1alpha1.BlockOutput, 0, 1)
	//		}
	//		blocks[lastBlock].Outputs = append(blocks[lastBlock].Outputs, &v1alpha1.BlockOutput{
	//			Items: []*v1alpha1.BlockOutputItem{
	//				{
	//					TextData: textData,
	//				},
	//			},
	//		})
	//	} else {
	//		block := &v1alpha1.Block{
	//			Kind:     v1alpha1.BlockKind_CODE,
	//			Contents: textData,
	//			Language: lang,
	//		}
	//		blocks = append(blocks, block)
	//	}
	//
	//	// We can skip walking the children of the code block since we've already ingested the code block
	//	return ast.WalkSkipChildren, nil
	//})
	//
	//// The way we walk the AST above we potentially end up segmenting continuous markdown without code blocks
	//// into more than one block. So we merge these blocks.
	//final := make([]*v1alpha1.Block, 0, len(blocks))
	//i := 0
	//for _, block := range blocks {
	//	lastBlock := i - 1
	//	addToLastBlock := false
	//	if lastBlock >= 0 && block.Kind == v1alpha1.BlockKind_MARKUP && final[lastBlock].Kind == v1alpha1.BlockKind_MARKUP {
	//		addToLastBlock = true
	//	}
	//
	//	if addToLastBlock {
	//		final[lastBlock].Contents += block.Contents
	//	} else {
	//		final = append(final, block)
	//		i++
	//	}
	//}
	//
	//return final, err
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

//func toCells(node *document.Node, source []byte) (result []*v1alpha1.Block) {
//	toCellsRec(node, &result, source)
//	return
//}
//
//func toCellsRec(
//	node *ast.Node,
//	cells *[]*v1alpha1.Block,
//	source []byte,
//) {
//	if node == nil {
//		return
//	}
//
//	for childIdx, child := range node.Children()
//		switch block := child.Item().(type) {
//		case *document.InnerBlock:
//			switch block.Unwrap().Kind() {
//			case ast.KindList:
//				nodeWithCode := document.FindNode(child, func(n *document.Node) bool {
//					return n.Item().Kind() == document.CodeBlockKind
//				})
//				if nodeWithCode == nil {
//					*cells = append(*cells, &Cell{
//						Kind:  MarkupKind,
//						Value: fmtValue(block.Value()),
//					})
//				} else {
//					for _, listItemNode := range child.Children() {
//						nodeWithCode := document.FindNode(listItemNode, func(n *document.Node) bool {
//							return n.Item().Kind() == document.CodeBlockKind
//						})
//						if nodeWithCode != nil {
//							toCellsRec(doc, listItemNode, cells, source)
//						} else {
//							*cells = append(*cells, &Cell{
//								Kind:  MarkupKind,
//								Value: fmtValue(listItemNode.Item().Value()),
//							})
//						}
//					}
//				}
//
//			case ast.KindBlockquote:
//				nodeWithCode := document.FindNode(child, func(n *document.Node) bool {
//					return n.Item().Kind() == document.CodeBlockKind
//				})
//				if nodeWithCode != nil {
//					toCellsRec(doc, child, cells, source)
//				} else {
//					*cells = append(*cells, &Cell{
//						Kind:  MarkupKind,
//						Value: fmtValue(block.Value()),
//					})
//				}
//			}
//
//		case *document.CodeBlock:
//			textRange := block.TextRange()
//
//			// In the future, we will include language detection (#77).
//			metadata := block.Attributes()
//			cellID := block.ID()
//			if cellID != "" {
//				metadata[PrefixAttributeName(InternalAttributePrefix, "id")] = cellID
//			}
//			metadata[PrefixAttributeName(InternalAttributePrefix, "name")] = block.Name()
//
//			nameGeneratedStr := "false"
//			if block.IsUnnamed() {
//				nameGeneratedStr = "true"
//			}
//			metadata[PrefixAttributeName(InternalAttributePrefix, "nameGenerated")] = nameGeneratedStr
//
//			*cells = append(*cells, &Cell{
//				Kind:       CodeKind,
//				Value:      string(block.Content()),
//				LanguageID: block.Language(),
//				Metadata:   metadata,
//				TextRange: &TextRange{
//					Start: textRange.Start + doc.ContentOffset(),
//					End:   textRange.End + doc.ContentOffset(),
//				},
//			})
//
//		case *document.MarkdownBlock:
//			value := block.Value()
//			astNode := block.Unwrap()
//
//			metadata := make(map[string]string)
//			_, includeAstMetadata := os.LookupEnv("RUNME_AST_METADATA")
//
//			if includeAstMetadata {
//				astMetadata := DumpToMap(astNode, source, astNode.Kind().String())
//				jsonAstMetaData, err := json.Marshal(astMetadata)
//				if err != nil {
//					log.Fatalf("Error converting to JSON: %s", err)
//				}
//
//				metadata["runme.dev/ast"] = string(jsonAstMetaData)
//			}
//
//			isListItem := node.Item() != nil && node.Item().Unwrap().Kind() == ast.KindListItem
//			if childIdx == 0 && isListItem {
//				listItem := node.Item().Unwrap().(*ast.ListItem)
//				list := listItem.Parent().(*ast.List)
//
//				var prefix []byte
//
//				if !list.IsOrdered() {
//					prefix = append(prefix, []byte{list.Marker, ' '}...)
//				} else {
//					itemNumber := list.Start
//					tmp := node.Item().Unwrap()
//					for tmp.PreviousSibling() != nil {
//						tmp = tmp.PreviousSibling()
//						itemNumber++
//					}
//					prefix = append([]byte(strconv.Itoa(itemNumber)), '.', ' ')
//				}
//
//				value = append(prefix, value...)
//			}
//
//			*cells = append(*cells, &Cell{
//				Kind:     MarkupKind,
//				Value:    fmtValue(value),
//				Metadata: metadata,
//			})
//		}
//	}
//}
