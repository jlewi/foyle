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

	test_dir := filepath.Join(cwd, "test_data", "logs")

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

	f, err := os.Open(name)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	d := json.NewDecoder(f)

	actual := map[string]*BlockLog{}
	for {
		var b BlockLog
		if err := d.Decode(&b); err != nil {
			if err == io.EOF {
				break
			}
			t.Errorf("Failed to decode block log: %v", err)
		}
		actual[b.ID] = &b
	}

	expectedBlocks := map[string]bool{
		"10b11f2d-7c8d-4d58-bedc-7e2dd51a85dc": true,
		"8594d742-39d7-473b-b4ad-901ff362fdb3": true,
		"f1662328-e884-418c-a084-95dfb1a3f7fc": true,
		"4feb6219-d050-4630-8ceb-d08ec149b60d": true,
		"3893a0b6-8c84-49ca-a38c-fbf6d7adfcde": true,
		"fd276a6f-f379-4f9c-9779-0ed07819d0f5": true,
		"d507ce35-af59-4f92-8dec-6c37d7b26647": true,
		"c885c6ba-598c-4fb1-8014-3cf8a330614c": true,
		"b56bb6bf-f631-4917-be48-ceeaf8797c41": true,
		"6e42f6f9-d394-41be-8baa-d7f8b41c0e11": true,
	}

	for id, _ := range expectedBlocks {
		if _, ok := actual[id]; !ok {
			t.Errorf("Missing block log for id ID: %v", id)
		}
	}

	for id, _ := range actual {
		if _, ok := expectedBlocks[id]; !ok {
			t.Errorf("Unexpected block log for id ID: %v", id)
		}
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
			entries := make([]*LogEntry, 0, 10)
			testFile, err := os.Open(filepath.Join(cwd, "test_data", c.linesFile))
			if err != nil {
				t.Fatalf("Failed to open test file: %v", err)
			}
			d := json.NewDecoder(testFile)
			for {
				e := &LogEntry{}
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
			entries := make([]*LogEntry, 0, 10)
			testFile, err := os.Open(filepath.Join(cwd, "test_data", c.linesFile))
			if err != nil {
				t.Fatalf("Failed to open test file: %v", err)
			}
			d := json.NewDecoder(testFile)
			for {
				e := &LogEntry{}
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
