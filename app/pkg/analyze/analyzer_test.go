package analyze

import (
	"context"
	"encoding/json"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jlewi/foyle/app/api"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jlewi/foyle/app/pkg/testutil"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"go.uber.org/zap"
)

func timeMustParse(layoutString, value string) time.Time {
	t, err := time.Parse(layoutString, value)
	if err != nil {
		panic(err)
	}
	return t
}

func shuffle(in []string) []string {
	rand.Shuffle(len(in), func(i, j int) { in[i], in[j] = in[j], in[i] })
	return in
}

func Test_BuildBlockLog(t *testing.T) {
	type testCase struct {
		name     string
		block    *api.BlockLog
		traces   map[string]api.Trace
		expected *api.BlockLog
	}

	traces := make(map[string]api.Trace)

	const bid1 = "g123output1"
	genTrace := &api.GenerateTrace{
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

	execTrace1 := &api.ExecuteTrace{
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

	execTrace2 := &api.ExecuteTrace{
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
			block: &api.BlockLog{
				ID:         bid1,
				GenTraceID: genTrace.TraceID,

				ExecTraceIDs: execTraceIds,
			},
			expected: &api.BlockLog{
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

	testDir := filepath.Join(cwd, "test_data", "logs")

	a, err := NewAnalyzer()
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	oDir, err := os.MkdirTemp("", "processedLogs")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	resultFiles, err := a.Analyze(context.Background(), testDir, oDir)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	t.Logf("Output written to: %s", oDir)

	// Check the blocks
	f, err := os.Open(resultFiles.BlockLogs[0])
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	d := json.NewDecoder(f)

	actual := map[string]*api.BlockLog{}
	for {
		var b api.BlockLog
		if err := d.Decode(&b); err != nil {
			if err == io.EOF {
				break
			}
			t.Errorf("Failed to decode block log: %v", err)
		}
		actual[b.ID] = &b
	}

	expectedBlocks := map[string]bool{
		"9557680b-e08c-4d1d-b098-6dcd03e0e108": true,
		"23706965-8e3b-440d-ba1a-1e1cc035fbd4": true,
		"48d530be-254a-493f-8cf4-20627078f830": true,
	}

	for id := range expectedBlocks {
		if _, ok := actual[id]; !ok {
			t.Errorf("Missing block log for id ID: %v", id)
		}
	}

	for id := range actual {
		if _, ok := expectedBlocks[id]; !ok {
			t.Errorf("Unexpected block log for id ID: %v", id)
		}
	}

	// This is a block that was generated via the AI and then executed so run some additional checks
	block := actual["23706965-8e3b-440d-ba1a-1e1cc035fbd4"]
	if block.GenTraceID == "" {
		t.Errorf("Expected GenTraceID to be set")
	}
	if len(block.ExecTraceIDs) == 0 {
		t.Errorf("Expected ExecTraceIDs to be set")
	}
	if block.Doc == nil {
		t.Errorf("Expected Doc to be set")
	}
	if block.GeneratedBlock == nil {
		t.Errorf("Expected GeneratedBlock to be set")
	}
	if block.ExecutedBlock == nil {
		t.Errorf("Expected ExecutedBlock to be set")
	}

	// Check the traces
	checkGenTracesFiles(t, resultFiles.GenerateTraces[0])
	checkExecuteTracesFiles(t, resultFiles.ExecuteTraces[0])
}

func checkGenTracesFiles(t *testing.T, path string) {
	// Check the generate traces
	genFile, err := os.Open(path)
	if err != nil {
		t.Errorf("Failed to open output file: %v", err)
		return
	}
	traces := make([]*api.GenerateTrace, 0, 10)
	d := json.NewDecoder(genFile)
	for {
		trace := &api.GenerateTrace{}
		if err := d.Decode(trace); err != nil {
			if err == io.EOF {
				break
			}
			t.Errorf("Failed to decode generate trace: %v", err)
		}
		traces = append(traces, trace)
	}

	if len(traces) == 0 {
		t.Errorf("Expected to find some generate traces")
	}
}

func checkExecuteTracesFiles(t *testing.T, path string) {
	// Check the generate traces
	genFile, err := os.Open(path)
	if err != nil {
		t.Errorf("Failed to open output file: %v", err)
		return
	}
	traces := make([]*api.ExecuteTrace, 0, 10)
	d := json.NewDecoder(genFile)
	for {
		trace := &api.ExecuteTrace{}
		if err := d.Decode(trace); err != nil {
			if err == io.EOF {
				break
			}
			t.Errorf("Failed to decode execute trace: %v", err)
		}
		traces = append(traces, trace)
	}

	if len(traces) == 0 {
		t.Errorf("Expected to find some execute traces")
	}
}

func Test_CombineGenerateEntries(t *testing.T) {
	type testCase struct {
		name      string
		linesFile string
	}

	cases := []testCase{
		{
			name:      "basic",
			linesFile: "generate_trace_lines.jsonl",
		},
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			entries := make([]*api.LogEntry, 0, 10)
			testFile, err := os.Open(filepath.Join(cwd, "test_data", c.linesFile))
			if err != nil {
				t.Fatalf("Failed to open test file: %v", err)
			}
			d := json.NewDecoder(testFile)
			for {
				e := &api.LogEntry{}
				err := d.Decode(e)
				if err != nil {
					if err == io.EOF {
						break
					}
					t.Fatalf("Failed to unmarshal log entry: %v", err)
				}
				entries = append(entries, e)
			}
			trace, err := combineGenerateTrace(context.Background(), entries)
			if err != nil {
				t.Fatalf("combineGenerateTrace failed: %+v", err)
			}
			if trace == nil {
				t.Fatalf("combineGenerateTrace should have returned non nil response")
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

func Test_CombineExecuteEntries(t *testing.T) {
	type testCase struct {
		name      string
		linesFile string
	}

	cases := []testCase{
		{
			name:      "basic",
			linesFile: "execute_traces_lines.jsonl",
		},
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			entries := make([]*api.LogEntry, 0, 10)
			testFile, err := os.Open(filepath.Join(cwd, "test_data", c.linesFile))
			if err != nil {
				t.Fatalf("Failed to open test file: %v", err)
			}
			d := json.NewDecoder(testFile)
			for {
				e := &api.LogEntry{}
				err := d.Decode(e)
				if err != nil {
					if err == io.EOF {
						break
					}
					t.Fatalf("Failed to unmarshal log entry: %v", err)
				}
				entries = append(entries, e)
			}
			trace, err := combineExecuteTrace(context.Background(), entries)
			if err != nil {
				t.Fatalf("combineExecuteTrace failed: %+v", err)
			}
			if trace == nil {
				t.Fatalf("combineExecuteTrace should have returned non nil response")
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
