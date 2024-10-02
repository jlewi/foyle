package analyze

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jlewi/foyle/app/pkg/logs/matchers"
	"github.com/sashabaranov/go-openai"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jlewi/foyle/app/api"
	"github.com/jlewi/foyle/app/pkg/testutil"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
)

func Test_logEntryToSpan(t *testing.T) {
	type testCase struct {
		name     string
		logLine  string
		expected *logspb.Span
	}

	cases := []testCase{
		{
			name:    "RAGSpan-Example",
			logLine: `{"severity":"info","time":1717094160.1880581,"caller":"learn/in_memory.go:104","function":"github.com/jlewi/foyle/app/pkg/learn.(*InMemoryExampleDB).GetExamples","message":"RAG result","traceId":"3fe82dae88bca105b92aee98c7f48228","evalMode":false,"example":{"id":"01HZ3K97HMF590J823F10RJZ4T","embedding":[],"query":{"blocks":[{"kind":"MARKUP","language":"markdown","contents":"Use gitops to aply the latest manifests to the dev cluster","outputs":[],"trace_ids":[],"id":""}]},"answer":[{"kind":"CODE","language":"","contents":"flux reconcile kustomization dev-cluster ----with-source ","outputs":[],"trace_ids":[],"id":""}]},"score":0.3000941151573202}`,
			expected: &logspb.Span{
				Data: &logspb.Span_Rag{
					Rag: &logspb.RAGSpan{
						Results: []*v1alpha1.RAGResult{
							{
								Example: &v1alpha1.Example{
									Id: "01HZ3K97HMF590J823F10RJZ4T",
									Query: &v1alpha1.Doc{
										Blocks: []*v1alpha1.Block{
											{
												Kind:     v1alpha1.BlockKind_MARKUP,
												Language: "markdown",
												Contents: "Use gitops to aply the latest manifests to the dev cluster",
											},
										},
									},

									Answer: []*v1alpha1.Block{
										{
											Kind:     v1alpha1.BlockKind_CODE,
											Contents: "flux reconcile kustomization dev-cluster ----with-source ",
										},
									},
								},
								Score: .3000941151573202,
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := &api.LogEntry{}
			if err := json.Unmarshal([]byte(tc.logLine), e); err != nil {
				t.Fatalf("Failed to unmarshal log line: %v", err)
			}
			span := logEntryToSpan(context.Background(), e)
			if d := cmp.Diff(tc.expected, span, cmpopts.IgnoreUnexported(logspb.Span{}, logspb.RAGSpan{}, v1alpha1.RAGResult{}, v1alpha1.Example{}), testutil.DocComparer); d != "" {
				t.Fatalf("Unexpected diff:\n%v", d)
			}
		})
	}
}

func Test_logEntrytoLLMSpan(t *testing.T) {
	type testCase struct {
		name     string
		entry    *api.LogEntry
		expected *logspb.Span
	}

	oaiRequestEntry := &api.LogEntry{
		"function": matchers.OAIComplete,
	}

	oaiRequest := &openai.ChatCompletionRequest{
		Model: "model",
	}

	if err := api.SetRequest(oaiRequestEntry, oaiRequest); err != nil {
		t.Fatalf("Failed to set request: %v", err)
	}

	oaiRequestJson, err := json.Marshal(oaiRequest)

	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	oaiResponseEntry := &api.LogEntry{
		"function": matchers.OAIComplete,
	}

	oaiResponse := &openai.ChatCompletionResponse{
		Model: "somemodel",
	}
	if err := api.SetResponse(oaiResponseEntry, oaiResponse); err != nil {
		t.Fatalf("Failed to set request: %v", err)
	}
	oaiResponseJson, err := json.Marshal(oaiResponse)

	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	cases := []testCase{
		{
			name:  "OAIRequest",
			entry: oaiRequestEntry,
			expected: &logspb.Span{
				Data: &logspb.Span_Llm{
					Llm: &logspb.LLMSpan{
						Provider:    v1alpha1.ModelProvider_OPEN_AI,
						RequestJson: string(oaiRequestJson),
					},
				},
			},
		},
		{
			name:  "OAIResponse",
			entry: oaiResponseEntry,
			expected: &logspb.Span{
				Data: &logspb.Span_Llm{
					Llm: &logspb.LLMSpan{
						Provider:     v1alpha1.ModelProvider_OPEN_AI,
						ResponseJson: string(oaiResponseJson),
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			span := logEntryToSpan(context.Background(), tc.entry)

			if span == nil {
				t.Fatalf("Expected non nil span")
			}

			if span.GetLlm() == nil {
				t.Fatalf("Expected LLM span")
			}

			if span.GetLlm().GetProvider() != v1alpha1.ModelProvider_OPEN_AI {
				t.Fatalf("Expected OpenAI provider")
			}

			// The JSON serialization of the proto isn't deterministic so we can't use cmp.diff to evaluate the response
			expectedHasRequest := tc.expected.GetLlm().GetRequestJson() != ""
			expectedHasResponse := tc.expected.GetLlm().GetResponseJson() != ""

			if expectedHasRequest {
				if span.GetLlm().GetRequestJson() == "" {
					t.Fatalf("Expected request")
				}
			}

			if expectedHasResponse {
				if span.GetLlm().GetResponseJson() == "" {
					t.Fatalf("Expected response")
				}
			}
		})

	}
}

func Test_CombineSpans(t *testing.T) {
	type testCase struct {
		name     string
		trace    *logspb.Trace
		expected *logspb.Trace
	}

	cases := []testCase{
		{
			name: "combine-ragspans",
			trace: &logspb.Trace{
				Spans: []*logspb.Span{
					{
						Data: &logspb.Span_Rag{
							Rag: &logspb.RAGSpan{
								Query: "query",
							},
						},
					},
					{
						Data: &logspb.Span_Rag{
							Rag: &logspb.RAGSpan{
								Results: []*v1alpha1.RAGResult{
									{
										Example: &v1alpha1.Example{
											Id: "012",
										},
									},
								},
							},
						},
					},
					{
						Data: &logspb.Span_Rag{
							Rag: &logspb.RAGSpan{
								Results: []*v1alpha1.RAGResult{
									{
										Example: &v1alpha1.Example{
											Id: "abc",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &logspb.Trace{
				Spans: []*logspb.Span{
					{
						Data: &logspb.Span_Rag{
							Rag: &logspb.RAGSpan{
								Query: "query",
								Results: []*v1alpha1.RAGResult{
									{
										Example: &v1alpha1.Example{
											Id: "012",
										},
									},
									{
										Example: &v1alpha1.Example{
											Id: "abc",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			combineSpans(tc.trace)
			if d := cmp.Diff(tc.expected, tc.trace, cmpopts.IgnoreUnexported(logspb.Trace{}, logspb.Span{}, logspb.RAGSpan{}, v1alpha1.RAGResult{}, v1alpha1.Example{}), testutil.DocComparer); d != "" {
				t.Fatalf("Unexpected diff:\n%v", d)
			}
		})
	}
}
