package analyze

import (
	"context"
	"strings"

	"github.com/jlewi/foyle/app/api"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
)

func logEntryToSpan(ctx context.Context, e *api.LogEntry) *logspb.Span {
	if strings.Contains(e.Function(), "learn.(*InMemoryExampleDB).GetExamples") {
		return logEntryToRAGSpan(ctx, e)
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
	for _, s := range oldSpans {
		if s.GetRag() != nil {
			if ragSpan == nil {
				ragSpan = s.GetRag()
			} else {
				ragSpan = combineRAGSpans(ragSpan, s.GetRag())
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
