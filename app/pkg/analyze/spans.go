package analyze

import (
	"context"
	"strings"

	"github.com/jlewi/foyle/app/pkg/logs/matchers"
	"google.golang.org/protobuf/proto"

	"github.com/jlewi/foyle/app/api"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
)

func logEntryToSpan(ctx context.Context, e *api.LogEntry) *logspb.Span {
	if strings.Contains(e.Function(), "learn.(*InMemoryExampleDB).GetExamples") {
		return logEntryToRAGSpan(ctx, e)
	}

	if matchers.IsOAIComplete(e.Function()) || matchers.IsAnthropicComplete(e.Function()) {
		return logEntryToLLMSpan(ctx, e)
	}
	return nil
}

func logEntryToLLMSpan(ctx context.Context, e *api.LogEntry) *logspb.Span {
	provider := v1alpha1.ModelProvider_MODEL_PROVIDER_UNKNOWN
	if matchers.IsOAIComplete(e.Function()) {
		provider = v1alpha1.ModelProvider_OPEN_AI
	} else if matchers.IsAnthropicComplete(e.Function()) {
		provider = v1alpha1.ModelProvider_ANTHROPIC
	}

	// Code relies on the fact that the completer field only use the fields request and response for the LLM model.
	// The code below also relies on the fact that the request and response are logged on separate log lines
	reqB := e.Request()
	if reqB != nil {
		return &logspb.Span{
			Data: &logspb.Span_Llm{
				Llm: &logspb.LLMSpan{
					Provider:    provider,
					RequestJson: string(reqB),
				},
			},
		}
	}

	resB := e.Response()
	if resB != nil {
		return &logspb.Span{
			Data: &logspb.Span_Llm{
				Llm: &logspb.LLMSpan{
					Provider:     provider,
					ResponseJson: string(resB),
				},
			},
		}
	}

	return nil
}

func logEntryToRAGSpan(ctx context.Context, e *api.LogEntry) *logspb.Span {
	rag := &logspb.RAGSpan{
		Results: make([]*v1alpha1.RAGResult, 0),
	}
	if v, ok := e.Get("query"); ok {
		if q, ok := v.(string); ok {
			rag.Query = q
		}
	}

	example := &v1alpha1.Example{}
	if e.GetProto("example", example) {
		// initialize score to a random negative value.
		// This will help us distinguish it being missing.
		score := -277.0
		if newScore, ok := e.GetFloat64("score"); ok {
			score = newScore
		}

		rag.Results = append(rag.Results, &v1alpha1.RAGResult{
			Example: example,
			Score:   score,
		})
	}

	return &logspb.Span{
		Data: &logspb.Span_Rag{
			Rag: rag,
		},
	}
}

// combineSpans within the trace
func combineSpans(trace *logspb.Trace) {
	oldSpans := trace.Spans
	trace.Spans = make([]*logspb.Span, 0, len(oldSpans))

	var ragSpan *logspb.RAGSpan
	var llmSpan *logspb.LLMSpan

	for _, s := range oldSpans {
		if s.GetRag() != nil {
			if ragSpan == nil {
				ragSpan = s.GetRag()
			} else {
				ragSpan = combineRAGSpans(ragSpan, s.GetRag())
			}
		} else if s.GetLlm() != nil {
			if llmSpan == nil {
				llmSpan = s.GetLlm()
			} else {
				llmSpan = combineLLMSpans(llmSpan, s.GetLlm())
			}
		} else {
			trace.Spans = append(trace.Spans, s)
		}
	}

	if ragSpan != nil {
		trace.Spans = append(trace.Spans, &logspb.Span{
			Data: &logspb.Span_Rag{
				Rag: ragSpan,
			},
		})
	}

	if llmSpan != nil {
		trace.Spans = append(trace.Spans, &logspb.Span{
			Data: &logspb.Span_Llm{
				Llm: llmSpan,
			},
		})
	}
}

// combine two RagSpans
func combineRAGSpans(a, b *logspb.RAGSpan) *logspb.RAGSpan {
	span := &logspb.RAGSpan{
		Query:   a.Query,
		Results: make([]*v1alpha1.RAGResult, 0, len(a.Results)+len(b.Results)),
	}

	if span.Query == "" && b.Query != "" {
		span.Query = b.Query
	}

	span.Results = append(span.Results, a.Results...)
	span.Results = append(span.Results, b.Results...)
	return span
}

// combine two LLMSpans
func combineLLMSpans(a, b *logspb.LLMSpan) *logspb.LLMSpan {
	proto.Merge(a, b)
	return a
}
