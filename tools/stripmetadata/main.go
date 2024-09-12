package main

import (
	"bytes"
	"context"
	"fmt"
	markdownfmt "github.com/Kunde21/markdownfmt/v3/markdown"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// findMDFiles returns a list of the all the markdown files in the eval directory.
func findMDFiles(ctx context.Context, evalDir string) ([]string, error) {
	examples := make([]string, 0, 100)
	err := filepath.Walk(evalDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".md" {
			return nil
		}

		examples = append(examples, path)
		return nil
	})

	return examples, err
}

func processFile(ctx context.Context, path string) error {
	input, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error reading input file:", err)
		return err
	}

	md := goldmark.New(
		goldmark.WithRenderer(
			markdownfmt.NewRenderer(),
		),
	)

	doc := md.Parser().Parse(text.NewReader(input))

	// Walk the AST and remove metadata from code blocks
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if cb, ok := n.(*ast.FencedCodeBlock); ok {
				// Remove all info strings (metadata)
				if cb.Info != nil {
					// Preserve only the language identifier (first word of Info)
					lang := strings.Fields(string(cb.Info.Text(input)))[0]
					cb.Info.Segment = text.NewSegment(cb.Info.Segment.Start, cb.Info.Segment.Start+len(lang))
				}
			}
		}
		return ast.WalkContinue, nil
	})

	var buf bytes.Buffer
	if err := md.Renderer().Render(&buf, input, doc); err != nil {
		return errors.Wrapf(err, "Error rendering markdown")
	}

	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return errors.Wrapf(err, "Error writing output file %v", path)
	}
	return nil
}

func run() error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.Wrapf(err, "Error getting current working directory")
	}

	mdFiles, err := findMDFiles(context.Background(), cwd)
	if err != nil {
		return errors.Wrapf(err, "Error finding markdown files")
	}

	for _, mdFile := range mdFiles {
		if err := processFile(context.Background(), mdFile); err != nil {
			return errors.Wrapf(err, "Error processing file %v", mdFile)
		}
	}
	return nil
}

func main() {
	processFile(context.Background(), "/Users/jlewi/git_foyle/docs/content/en/docs/learning/troubleshoot_learning.md")
	if err := run(); err != nil {
		fmt.Println("Error processing markdown: %+v", err)
		os.Exit(1)
	}

	fmt.Println("Markdown processed successfully!")
}
