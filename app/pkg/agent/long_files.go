package agent

import (
	"bufio"
	_ "embed"
	"github.com/jlewi/monogo/helpers"
	"os"
	"strings"
	"text/template"
)

//go:embed edit_file_prompt.tmpl
var editPrompt string

var (
	editTemplate = template.Must(template.New("prompt").Parse(editPrompt))
)

type editPromptInput struct {
	Changes string
	Text    string
}

// FileSnippet holds the start and end lines and the actual text of the segment.
type FileSnippet struct {
	StartLine int
	EndLine   int
	Text      string
}

// ReadFileSegment reads the file at filePath and returns the lines between startLine and endLine (inclusive) as a FileSegment.
func ReadFileSegment(filePath string, startLine, endLine int) (FileSnippet, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return FileSnippet{}, err
	}
	defer helpers.DeferIgnoreError(file.Close)

	scanner := bufio.NewScanner(file)
	var textBuilder strings.Builder
	currentLine := 0

	for scanner.Scan() {
		currentLine++
		if currentLine >= startLine && currentLine <= endLine {
			textBuilder.WriteString(scanner.Text() + "\n")
		}
		if currentLine > endLine {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return FileSnippet{}, err
	}

	return FileSnippet{
		StartLine: startLine,
		EndLine:   endLine,
		Text:      textBuilder.String(),
	}, nil
}
