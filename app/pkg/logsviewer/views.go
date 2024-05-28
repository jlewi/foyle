package logsviewer

import (
	"bytes"

	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"

	"github.com/jlewi/foyle/app/pkg/docs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"

	"github.com/go-logr/zapr"
	"github.com/pkg/errors"
	"github.com/yuin/goldmark"
	"go.uber.org/zap"
)

// renderGeneratedBlock returns the generated block as HTML if there is one
func renderGeneratedBlock(block *logspb.BlockLog) (string, error) {
	if block == nil {
		return "", errors.New("block is nil")
	}
	log := zapr.NewLogger(zap.L())

	if block.GeneratedBlock == nil {
		return "Block was not generated by the assistant", nil
	}

	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(block.GeneratedBlock.Contents), &buf); err != nil {
		log.Error(err, "Failed to convert markdown")
		return "", err
	}

	return buf.String(), nil
}

// renderExecutedBlock returns the executed block as HTML if there is one
func renderExecutedBlock(block *logspb.BlockLog) (string, error) {
	if block == nil {
		return "", errors.New("block is nil")
	}
	log := zapr.NewLogger(zap.L())

	if block.ExecutedBlock == nil {
		return "Block was not executed", nil
	}

	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(block.ExecutedBlock.Contents), &buf); err != nil {
		log.Error(err, "Failed to convert markdown")
		return "", err
	}

	return buf.String(), nil
}

// docToHTML returns the dock as html
func docToHTML(doc *v1alpha1.Doc) (string, error) {
	if doc == nil {
		return "", errors.New("doc is nil")
	}
	log := zapr.NewLogger(zap.L())

	// Convert it to markdown
	md := docs.DocToMarkdown(doc)

	// Conver the markdown to html
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(md), &buf); err != nil {
		log.Error(err, "Failed to convert markdown")
		return "", err
	}

	return buf.String(), nil
}
