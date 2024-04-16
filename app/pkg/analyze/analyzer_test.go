package analyze

import (
	"context"
	"encoding/json"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jlewi/foyle/app/pkg/testutil"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"go.uber.org/zap"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func timeMustParse(layoutString, value string) time.Time {
	t, err := time.Parse(layoutString, value)
	if err != nil {
		panic(err)
	}
	return t
}

func shuffle(in []string) []string {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(in), func(i, j int) { in[i], in[j] = in[j], in[i] })
	return in
}

func Test_BuildBlockLog(t *testing.T) {
	type testCase struct {
		name     string
		block    *BlockLog
		traces   map[string]Trace
		expected *BlockLog
	}

	traces := make(map[string]Trace)

	const bid1 = "g123output1"
	genTrace := &GenerateTrace{
		TraceID:   "g123",
		StartTime: timeMustParse(time.RFC3339, "2021-01-01T00:00:00Z"),
		EndTime:   timeMustParse(time.RFC3339, "2021-01-01T00:01:00Z"),
		Request: &v1alpha1.GenerateRequest{
			Doc: &v1alpha1.Doc{
				Blocks: []*v1alpha1.Block{
					{
						Contents: "echo hello",
					},
				},
			},
		},
		Response: &v1alpha1.GenerateResponse{
			Blocks: []*v1alpha1.Block{
				{
					Id:       bid1,
					Contents: "outcell",
				},
			},
		},
	}

	execTrace1 := &ExecuteTrace{
		TraceID:   "e456",
		StartTime: timeMustParse(time.RFC3339, "2021-01-02T00:00:00Z"),
		EndTime:   timeMustParse(time.RFC3339, "2021-01-02T00:01:00Z"),
		Request: &v1alpha1.ExecuteRequest{
			Block: &v1alpha1.Block{
				Contents: "echo hello",
				Id:       bid1,
			},
		},
		Response: &v1alpha1.ExecuteResponse{
			Outputs: []*v1alpha1.BlockOutput{
				{
					Items: []*v1alpha1.BlockOutputItem{
						{
							TextData: "exitCode: 4",
						},
					},
				},
			},
		},
	}

	execTrace2 := &ExecuteTrace{
		TraceID:   "e789",
		StartTime: timeMustParse(time.RFC3339, "2021-01-03T00:00:00Z"),
		EndTime:   timeMustParse(time.RFC3339, "2021-01-03T00:01:00Z"),
		Request: &v1alpha1.ExecuteRequest{
			Block: &v1alpha1.Block{
				Contents: "echo hello",
				Id:       bid1,
			},
		},
		Response: &v1alpha1.ExecuteResponse{
			Outputs: []*v1alpha1.BlockOutput{
				{
					Items: []*v1alpha1.BlockOutputItem{
						{
							TextData: "exitCode: 7",
						},
					},
				},
			},
		},
	}

	traces[genTrace.TraceID] = genTrace
	traces[execTrace1.TraceID] = execTrace1
	traces[execTrace2.TraceID] = execTrace2

	// We shuffle ExecTraceIds to make sure we properly set block log based on the later trace
	execTraceIds := shuffle([]string{execTrace1.TraceID, execTrace2.TraceID})
	cases := []testCase{
		{
			name: "basic",
			block: &BlockLog{
				ID:         bid1,
				GenTraceID: genTrace.TraceID,

				ExecTraceIDs: execTraceIds,
			},
			expected: &BlockLog{
				ID:             bid1,
				GenTraceID:     genTrace.TraceID,
				ExecTraceIDs:   execTraceIds,
				Doc:            genTrace.Request.Doc,
				GeneratedBlock: genTrace.Response.Blocks[0],
				ExecutedBlock:  execTrace2.Request.Block,
				ExitCode:       7,
			},
			traces: traces,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := buildBlockLog(context.Background(), c.block, c.traces); err != nil {
				t.Fatalf("buildBlockLog failed: %v", err)
			}

			if d := cmp.Diff(c.expected, c.block, testutil.BlockComparer, cmpopts.IgnoreUnexported(v1alpha1.Doc{})); d != "" {
				t.Errorf("Unexpected diff:\n%s", d)
			}
		})
	}
}

func Test_Analyzer(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	c := zap.NewDevelopmentConfig()
	log, err := c.Build()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	zap.ReplaceGlobals(log)

	test_dir := filepath.Join(cwd, "test_data")

	a, err := NewAnalyzer()
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	oFile, err := os.CreateTemp("", "output.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	name := oFile.Name()
	if err := oFile.Close(); err != nil {
		t.Fatalf("Failed to close file: %v", err)
	}

	if err := a.Analyze(context.Background(), test_dir, name); err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	t.Logf("Output written to: %s", name)

	actual := make([]*BlockLog, 0, 10)
	f, err := os.Open(name)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	d := json.NewDecoder(f)

	for {
		var b BlockLog
		if err := d.Decode(&b); err != nil {
			if err == io.EOF {
				break
			}
			t.Errorf("Failed to decode block log: %v", err)
		}
		actual = append(actual, &b)
	}

	if len(actual) != 1 {
		t.Errorf("Expected 1 block log but got: %v", len(actual))
	}
}

func Test_CombineGenerateEntries(t *testing.T) {
	type testCase struct {
		name     string
		logLines []string
		expected []*GenerateTrace
	}

	cases := []testCase{
		{
			name: "basic",
			logLines: []string{
				`{"severity":"info","time":1713277485.655751,"caller":"agent/agent.go:61","function":"github.com/jlewi/foyle/app/pkg/agent.(*Agent).Generate","message":"Agent.Generate","traceId":"1e2720dfac7f5b810d6a5e230609cfc8","request":"doc:{blocks:{kind:MARKUP  language:\"markdown\"  contents:\"Use gcloud to read the logs for the cluster dev in project foyle-dev\"  trace_ids:\"\"  trace_ids:\"\"  trace_ids:\"\"}  blocks:{kind:MARKUP  contents:\"To read the logs for the cluster\"  id:\"1f051c78-1308-4fb0-9b09-7f120a8f26ad\"}  blocks:{kind:CODE  language:\"bash\"  contents:\"gcloud logging read\\n\"  id:\"a50a90ef-c396-45b7-8840-e4de9e3abf20\"}}"}`,
				`{"severity":"info","time":1713277485.656407,"caller":"agent/agent.go:116","function":"github.com/jlewi/foyle/app/pkg/agent.(*Agent).completeWithRetries","message":"OpenAI:CreateChatCompletion","traceId":"1e2720dfac7f5b810d6a5e230609cfc8","request":{"model":"gpt-3.5-turbo-0125","messages":[{"role":"system","content":"You are a helpful AI assistant for software developers. You are helping software engineers write markdown documents to deploy\nand operate software. Your job is to help users reason about problems and tasks and come up with the appropriate\ncommands to accomplish them. You should never try to execute commands. You should always tell the user\nto execute the commands themselves. To help the user place the commands inside a code block with the language set to\nbash. Users can then execute the commands inside VSCode notebooks. The output will then be appended to the document.\nYou can then use that output to reason about the next steps.\n\nYou are only helping users with tasks related to building, deploying, and operating software. You should interpret\nany questions or commands in that context.\n"},{"role":"user","content":"openai response"}],"max_tokens":2000,"temperature":0.9}}`,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			entries := make([]*LogEntry, 0, len(c.logLines))
			for _, l := range c.logLines {
				e := &LogEntry{}
				if err := json.Unmarshal([]byte(l), e); err != nil {
					t.Fatalf("Failed to unmarshal log entry: %v", err)
				}
				entries = append(entries, e)
			}
			trace, err := combineGenerateTrace(context.Background(), entries)
			if err != nil {
				t.Fatalf("combineEntriesForTrace failed: %v", err)
			}
			if trace == nil {
				t.Fatalf("combineEntriesForTrace should have returned non nil response")
			}

			// Assert the trace has a request and a response
			if trace.Request == nil {
				t.Errorf("Expected trace to have a request")
			}
			if trace.Response == nil {
				t.Errorf("Expected trace to have a response")
			}
		})
	}
}
