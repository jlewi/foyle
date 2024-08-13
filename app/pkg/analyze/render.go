package analyze

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/go-logr/zapr"
	"github.com/liushuangls/go-anthropic/v2"
	"github.com/yuin/goldmark"
	"go.uber.org/zap"
	"html/template"
)

//go:embed request.html.tmpl
var requestTemplateRaw string

var requestTemplate *template.Template

type TemplateData struct {
	Model       string
	Tokens      int
	Temperature float64
	System      string
	Messages    []Message
}

type Message struct {
	Role    string
	Content template.HTML
}

// renderAnthropicRequest returns a string containing the HTML representation of the request
func renderAnthropicRequest(request *anthropic.MessagesRequest) string {
	log := zapr.NewLogger(zap.L())
	data := &TemplateData{
		Model:       request.Model,
		Tokens:      request.MaxTokens,
		Temperature: float64(*request.Temperature),
		System:      request.System,
		Messages:    make([]Message, 0, len(request.Messages)),
	}

	for _, message := range request.Messages {
		content := ""
		for _, c := range message.Content {
			content += *c.Text
		}

		var buf bytes.Buffer
		if err := goldmark.Convert([]byte(content), &buf); err != nil {
			log.Error(err, "Failed to convert markdown to HTML")
			buf.WriteString(fmt.Sprintf("Failed to convert markdown to HTML: error %+v", err))
		}
		data.Messages = append(data.Messages, Message{
			Role:    message.Role,
			Content: template.HTML(buf.String()),
		})
	}
	var buf bytes.Buffer
	if err := requestTemplate.Execute(&buf, data); err != nil {
		log.Error(err, "Failed to execute request template")
		return fmt.Sprintf("Failed to execute request template: error %+v", err)
	}
	return buf.String()
}

func init() {
	// Register the template functions for the request template
	// This is necessary to be able to render the request template
	// with the data we pass to it
	var err error
	requestTemplate, err = template.New("request").Parse(requestTemplateRaw)
	if err != nil {
		panic(err)
	}
}
