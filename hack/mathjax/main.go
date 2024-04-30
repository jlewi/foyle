// mathjax is a hack to render a markdown file with mathjax support and open it in the browser
package main

import (
	"bytes"
	"fmt"
	mathjax "github.com/litao91/goldmark-mathjax"
	"github.com/pkg/browser"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"os"
	"path/filepath"
	"time"
)

func main() {
	var cmd = &cobra.Command{
		Use:   "mathjax [file]",
		Short: "Math jax converts the markdown file to html with mathjax support and then opens it up",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := convert(args[0]); err != nil {
				fmt.Printf("Failed to convert document: %+v", err)
			}
		},
	}

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

func convert(mdFile string) error {
	md := goldmark.New(
		goldmark.WithExtensions(mathjax.MathJax),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	mdContent, err := os.ReadFile(mdFile)
	if err != nil {
		return errors.Wrapf(err, "Failed to read file %s", mdFile)
	}
	// todo more control on the parsing process
	var html bytes.Buffer

	if err := md.Convert(mdContent, &html); err != nil {
		fmt.Println(err)
	}
	tdir := os.TempDir()

	tStamp := time.Now().Format("20060102-150405")
	oFile := filepath.Join(tdir, fmt.Sprintf("mathjax-%s.html", tStamp))

	f, err := os.Create(oFile)
	if err != nil {
		return errors.Wrap(err, "Failed to create temp file")
	}

	f.WriteString("<html>\n")
	f.WriteString("<head>\n")
	f.WriteString(`<script type="text/javascript" id="MathJax-script" async src="https://cdn.jsdelivr.net/npm/mathjax@3/es5/tex-mml-chtml.js"> </script>\n`)
	f.WriteString("</head>\n")
	if _, err := f.Write(html.Bytes()); err != nil {
		return errors.Wrap(err, "Failed to write to temp file")
	}
	f.WriteString("</html>")
	if err := f.Close(); err != nil {
		return errors.Wrap(err, "Failed to close temp file")
	}
	if err := browser.OpenFile(oFile); err != nil {
		return errors.Wrap(err, "Failed to open browser")
	}
	return nil
}
