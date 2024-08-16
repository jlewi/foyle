package agent

import (
	_ "embed"
	"text/template"
)

const (
	systemPrompt = `You are a helpful AI assistant for software developers. You are helping software engineers write 
markdown documents to deploy and operate software. Your job is to help users with tasks related to building, deploying,
and operating software. You should interpret any questions or commands in that context. You job is to suggest
commands the user can execute to accomplish their goals.`
)

//go:embed prompt.tmpl
var promptTemplateString string

var (
	promptTemplate = template.Must(template.New("prompt").Parse(promptTemplateString))
)

type Example struct {
	Input  string
	Output string
}

type promptArgs struct {
	Document string
	Examples []Example
}
