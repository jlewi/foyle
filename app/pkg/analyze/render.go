package analyze

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"

	"github.com/go-logr/zapr"
	"github.com/liushuangls/go-anthropic/v2"
	"github.com/yuin/goldmark"
	"go.uber.org/zap"
)

//go:embed request.html.tmpl
var requestTemplateRaw string

//go:embed response.html.tmpl
var responseTemplateRaw string

var requestTemplate *template.Template
var responseTemplate *template.Template

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

type ResponseTemplateData struct {
	ID           string
	Type         string
	Role         string
	Messages     []Message
	Model        string
	StopReason   string
	StopSequence string
	InputTokens  int
	OutputTokens int
}

// renderAnthropicResponse returns a string containing the HTML representation of the response
func renderAnthropicResponse(resp *anthropic.MessagesResponse) string {
	log := zapr.NewLogger(zap.L())
	data := &ResponseTemplateData{
		ID:           resp.ID,
		Type:         string(resp.Type),
		Role:         resp.Role,
		Model:        resp.Model,
		StopReason:   string(resp.StopReason),
		StopSequence: resp.StopSequence,
		InputTokens:  resp.Usage.InputTokens,
		OutputTokens: resp.Usage.OutputTokens,
		Messages:     make([]Message, 0, len(resp.Content)),
	}

	for _, message := range resp.Content {
		var buf bytes.Buffer
		if err := goldmark.Convert([]byte(message.GetText()), &buf); err != nil {
			log.Error(err, "Failed to convert markdown to HTML")
			buf.WriteString(fmt.Sprintf("Failed to convert markdown to HTML: error %+v", err))
		}
		data.Messages = append(data.Messages, Message{
			Content: template.HTML(buf.String()),
		})
	}
	var buf bytes.Buffer
	if err := responseTemplate.Execute(&buf, data); err != nil {
		log.Error(err, "Failed to execute response template")
		return fmt.Sprintf("Failed to execute response template: error %+v", err)
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

	responseTemplate, err = template.New("response").Parse(responseTemplateRaw)
	if err != nil {
		panic(err)
	}
}
