package analyze

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	"github.com/jlewi/foyle/app/api"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"

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

// renderHTML populates the HTML fields in GetLLMLogsResponse with the rendered HTML.
// The HTML is generated from the JSON of the request and response
func renderHTML(resp *logspb.GetLLMLogsResponse, provider api.ModelProvider) error {
	if resp == nil {
		return errors.WithStack(errors.New("response is nil"))
	}

	reqHtml, reqErr := RenderRequestHTML(resp.GetRequestJson(), provider)
	if reqErr != nil {
		return reqErr
	}
	resp.RequestHtml = reqHtml

	respHtml, respErr := RenderResponseHTML(resp.GetResponseJson(), provider)
	if respErr != nil {
		return respErr
	}
	resp.ResponseHtml = respHtml
	return nil
}

func RenderRequestHTML(jsonValue string, provider api.ModelProvider) (string, error) {
	if jsonValue == "" {
		return "", errors.WithStack(errors.New("request is empty"))
	}

	var data *TemplateData
	switch provider {
	case api.ModelProviderOpenAI:
		req := &openai.ChatCompletionRequest{}
		if err := json.Unmarshal([]byte(jsonValue), req); err != nil {
			return "", errors.Wrapf(err, "failed to unmarshal request to openai.ChatCompletionRequest; json: %s", jsonValue)
		}
		data = oaiRequestToTemplateData(req)
	case api.ModelProviderAnthropic:
		req := &anthropic.MessagesRequest{}
		if err := json.Unmarshal([]byte(jsonValue), req); err != nil {
			return "", errors.Wrapf(err, "failed to unmarshal request to anthropic.MessagesRequest; json: %s", jsonValue)
		}
		data = anthropicRequestToTemplateData(req)
	default:
		return fmt.Sprintf("<html><body><h1>Unsupported provider: %v</h1></body></html>", provider), nil
	}

	var buf bytes.Buffer
	if err := requestTemplate.Execute(&buf, data); err != nil {
		return "", errors.Wrapf(err, "Failed to execute request template")
	}

	return buf.String(), nil
}

func RenderResponseHTML(jsonValue string, provider api.ModelProvider) (string, error) {
	if jsonValue == "" {
		return "", errors.WithStack(errors.New("response is nil"))
	}

	var data *ResponseTemplateData
	switch provider {
	case api.ModelProviderOpenAI:
		req := &openai.ChatCompletionResponse{}
		if err := json.Unmarshal([]byte(jsonValue), req); err != nil {
			return "", errors.Wrapf(err, "failed to unmarshal request to openai.ChatCompletionResponse; json: %s", jsonValue)
		}
		data = oaiResponseToTemplateData(req)
	case api.ModelProviderAnthropic:
		req := &anthropic.MessagesResponse{}
		if err := json.Unmarshal([]byte(jsonValue), req); err != nil {
			return "", errors.Wrapf(err, "failed to unmarshal request to anthropic.MessagesRequest; json: %s", jsonValue)
		}
		data = anthropicResponseToTemplateData(req)
	default:
		return fmt.Sprintf("<html><body><h1>Unsupported provider: %v</h1></body></html>", provider), nil
	}

	var buf bytes.Buffer
	if err := responseTemplate.Execute(&buf, data); err != nil {
		return "", errors.Wrapf(err, "Failed to execute request template")
	}

	return buf.String(), nil
}

func anthropicRequestToTemplateData(request *anthropic.MessagesRequest) *TemplateData {
	log := zapr.NewLogger(zap.L())
	data := &TemplateData{
		Model:       request.Model,
		Tokens:      request.MaxTokens,
		Temperature: float64(*request.Temperature),
		System:      request.System,
		Messages:    make([]Message, 0, len(request.Messages)),
	}

	md := converter()

	for _, message := range request.Messages {
		content := ""
		for _, c := range message.Content {
			content += *c.Text
		}
		content = escapePromptTags(content)
		var buf bytes.Buffer
		if err := md.Convert([]byte(content), &buf); err != nil {
			log.Error(err, "Failed to convert markdown to HTML")
			buf.WriteString(fmt.Sprintf("Failed to convert markdown to HTML: error %+v", err))
		}
		data.Messages = append(data.Messages, Message{
			Role:    message.Role,
			Content: template.HTML(buf.String()),
		})
	}
	return data
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

func anthropicResponseToTemplateData(resp *anthropic.MessagesResponse) *ResponseTemplateData {
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
		content := message.GetText()
		content = escapePromptTags(content)
		var buf bytes.Buffer
		if err := goldmark.Convert([]byte(content), &buf); err != nil {
			log.Error(err, "Failed to convert markdown to HTML")
			buf.WriteString(fmt.Sprintf("Failed to convert markdown to HTML: error %+v", err))
		}
		data.Messages = append(data.Messages, Message{
			Content: template.HTML(buf.String()),
		})
	}
	return data
}

func oaiRequestToTemplateData(request *openai.ChatCompletionRequest) *TemplateData {
	log := zapr.NewLogger(zap.L())
	data := &TemplateData{
		Model:       request.Model,
		Tokens:      request.MaxTokens,
		Temperature: float64(request.Temperature),
		// System message for OpenAI is just a message with a role of system prompt
		System:   "",
		Messages: make([]Message, 0, len(request.Messages)),
	}

	md := converter()
	for _, message := range request.Messages {
		content := message.Content
		content = escapePromptTags(content)
		var buf bytes.Buffer
		if err := md.Convert([]byte(content), &buf); err != nil {
			log.Error(err, "Failed to convert markdown to HTML")
			buf.WriteString(fmt.Sprintf("Failed to convert markdown to HTML: error %+v", err))
		}
		data.Messages = append(data.Messages, Message{
			Role:    message.Role,
			Content: template.HTML(buf.String()),
		})
	}

	return data
}

func oaiResponseToTemplateData(resp *openai.ChatCompletionResponse) *ResponseTemplateData {
	log := zapr.NewLogger(zap.L())

	data := &ResponseTemplateData{
		ID:           resp.ID,
		Type:         "",
		Role:         "",
		Model:        resp.Model,
		StopSequence: "",
		InputTokens:  resp.Usage.PromptTokens,
		OutputTokens: resp.Usage.CompletionTokens,
		Messages:     make([]Message, 0, len(resp.Choices)),
	}

	if len(resp.Choices) > 0 {
		data.StopReason = string(resp.Choices[0].FinishReason)
		data.Role = resp.Choices[0].Message.Role
	}

	md := converter()
	for _, c := range resp.Choices {
		var buf bytes.Buffer
		content := escapePromptTags(c.Message.Content)
		if err := md.Convert([]byte(content), &buf); err != nil {
			log.Error(err, "Failed to convert markdown to HTML")
			buf.WriteString(fmt.Sprintf("Failed to convert markdown to HTML: error %+v", err))
		}
		data.Messages = append(data.Messages, Message{
			Content: template.HTML(buf.String()),
		})
	}
	return data
}

func converter() goldmark.Markdown {
	md := goldmark.New()
	return md
}

// escapePromptTags escapes the xml tags we use in our prompt so they can be displayed in the HTML
func escapePromptTags(data string) string {
	tags := []string{"<example>", "</example>", "<input>", "</input>", "<output>", "</output>"}

	for _, tag := range tags {
		escapedTag := tag
		escapedTag = strings.ReplaceAll(escapedTag, "<", "&lt;")
		escapedTag = strings.ReplaceAll(escapedTag, ">", "&gt;")

		data = strings.ReplaceAll(data, tag, escapedTag)
	}
	return data
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
