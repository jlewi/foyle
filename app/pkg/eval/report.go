package eval

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"

	"github.com/jlewi/foyle/app/api"
	"github.com/jlewi/foyle/app/pkg/analyze"
	"github.com/jlewi/foyle/app/pkg/docs"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/app/pkg/runme/converters"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/pkg/errors"
	parserv1 "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/parser/v1"
	"github.com/yuin/goldmark"
)

//go:embed report.html.tmpl
var reportTemplateRaw string

var reportTemplate *template.Template

// buildEvalReport builds an HTML report to display the evaluation results.
func buildEvalReport(ctx context.Context, result *v1alpha1.EvalResult, trace *logspb.Trace) (string, error) {
	log := logs.FromContext(ctx)
	data := &reportData{
		Result: result,
		Trace:  trace,
	}

	var expectedCells bytes.Buffer

	expectedDoc, err := converters.NotebookToDoc(&parserv1.Notebook{
		Cells: result.Example.ExpectedCells,
	})
	if err != nil {
		return "", errors.Wrapf(err, "Failed to convert notebook to doc")

	}

	data.JudgeExplanationHTML = explanationToHTML(result.JudgeExplanation)
	expectedMD := docs.DocToMarkdown(expectedDoc)

	if err := goldmark.Convert([]byte(expectedMD), &expectedCells); err != nil {
		log.Error(err, "Failed to convert markdown to HTML")
	} else {
		data.ExpectedResponseHTML = template.HTML(expectedCells.String())
	}

	data.LLMRequestHTML, data.LLMResponseHTML = llmSpanToHTML(ctx, trace)

	var buf bytes.Buffer
	if err := reportTemplate.Execute(&buf, data); err != nil {
		return "", errors.Wrapf(err, "Failed to execute request template")
	}

	return buf.String(), nil
}

func llmSpanToHTML(ctx context.Context, trace *logspb.Trace) (template.HTML, template.HTML) {
	if trace == nil {
		return template.HTML("No generate trace was provided"), template.HTML("No generate trace was provided")
	}

	var llmSpan *logspb.LLMSpan
	for _, t := range trace.GetSpans() {
		if t.GetLlm() != nil {
			llmSpan = t.GetLlm()
			break
		}
	}

	if llmSpan == nil {
		return template.HTML("No LLM span was found"), template.HTML("No LLM span was found")
	}

	var requestHTML template.HTML
	var responseHTML template.HTML
	requestString, err := analyze.RenderRequestHTML(llmSpan.RequestJson, api.ModelProviderProtoToAPI(llmSpan.GetProvider()))

	if err != nil {
		requestHTML = template.HTML(fmt.Sprintf("Failed to render LLM request; error: %+v", err))
	} else {
		requestHTML = template.HTML(requestString)
	}

	responseString, err := analyze.RenderResponseHTML(llmSpan.ResponseJson, api.ModelProviderProtoToAPI(llmSpan.GetProvider()))

	if err != nil {
		responseHTML = template.HTML(fmt.Sprintf("Failed to render LLM response; error: %+v", err))
	} else {
		responseHTML = template.HTML(responseString)
	}

	return requestHTML, responseHTML
}

func explanationToHTML(explanation string) template.HTML {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(explanation), &buf); err != nil {
		return template.HTML(fmt.Sprintf("<pre>%s</pre><br><br> **Note** Judge explanation could not be converted to HTML; error: %+v", explanation, err))
	}

	return template.HTML(buf.String())
}

type reportData struct {
	Result               *v1alpha1.EvalResult
	Trace                *logspb.Trace
	LLMResponseHTML      template.HTML
	ExpectedResponseHTML template.HTML
	LLMRequestHTML       template.HTML
	JudgeExplanationHTML template.HTML
}

func init() {
	// Register the template functions.
	var err error
	reportTemplate, err = template.New("report").Parse(reportTemplateRaw)
	if err != nil {
		panic(err)
	}
}
