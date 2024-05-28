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

	runnerv1 "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/runner/v1"

	"github.com/cockroachdb/pebble"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jlewi/foyle/app/pkg/dbutil"
	"github.com/jlewi/foyle/app/pkg/testutil"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/jlewi/monogo/helpers"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/jlewi/foyle/app/api"
)

func timeMustParse(layoutString, value string) *timestamppb.Timestamp {
	t, err := time.Parse(layoutString, value)
	if err != nil {
		panic(err)
	}
	return timestamppb.New(t)
}

func shuffle(in []string) []string {
	rand.Shuffle(len(in), func(i, j int) { in[i], in[j] = in[j], in[i] })
	return in
}

func Test_BuildBlockLog(t *testing.T) {
	type testCase struct {
		name     string
		block    *logspb.BlockLog
		traces   map[string]*logspb.Trace
		expected *logspb.BlockLog
	}

	traces := make(map[string]*logspb.Trace)

	const bid1 = "g123output1"
	genTrace := &logspb.Trace{
		Id:        "g123",
		StartTime: timeMustParse(time.RFC3339, "2021-01-01T00:00:00Z"),
		EndTime:   timeMustParse(time.RFC3339, "2021-01-01T00:01:00Z"),
		Data: &logspb.Trace_Generate{
			Generate: &logspb.GenerateTrace{
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
			},
		},
	}

	execTrace1 := &logspb.Trace{
		Id:        "e456",
		StartTime: timeMustParse(time.RFC3339, "2021-01-02T00:00:00Z"),
		EndTime:   timeMustParse(time.RFC3339, "2021-01-02T00:01:00Z"),
		Data: &logspb.Trace_Execute{
			Execute: &logspb.ExecuteTrace{
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
			},
		},
	}

	execTrace2 := &logspb.Trace{
		Id:        "e789",
		StartTime: timeMustParse(time.RFC3339, "2021-01-03T00:00:00Z"),
		EndTime:   timeMustParse(time.RFC3339, "2021-01-03T00:01:00Z"),
		Data: &logspb.Trace_Execute{
			Execute: &logspb.ExecuteTrace{
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
			},
		},
	}

	// Create a block in evaluation mode
	const bid2 = "g456output1"
	genTrace2 := &logspb.Trace{
		Id:        "g456",
		StartTime: timeMustParse(time.RFC3339, "2021-01-01T00:00:00Z"),
		EndTime:   timeMustParse(time.RFC3339, "2021-01-01T00:01:00Z"),
		Data: &logspb.Trace_Generate{
			Generate: &logspb.GenerateTrace{
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
							Id:       bid2,
							Contents: "outcell",
						},
					},
				},
			},
		},
		EvalMode: true,
	}

	execTrace3 := &logspb.Trace{
		Id:        "e912",
		StartTime: timeMustParse(time.RFC3339, "2021-01-03T00:00:00Z"),
		EndTime:   timeMustParse(time.RFC3339, "2021-01-03T00:01:00Z"),
		Data: &logspb.Trace_Execute{
			Execute: &logspb.ExecuteTrace{
				Request: &v1alpha1.ExecuteRequest{
					Block: &v1alpha1.Block{
						Contents: "echo hello",
						Id:       bid2,
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
			},
		},
	}

	traces[genTrace.Id] = genTrace
	traces[genTrace2.Id] = genTrace2
	traces[execTrace1.Id] = execTrace1
	traces[execTrace2.Id] = execTrace2
	traces[execTrace3.Id] = execTrace3

	// We shuffle ExecTraceIds to make sure we properly set block log based on the later trace
	execTraceIds := shuffle([]string{execTrace1.GetId(), execTrace2.GetId()})
	cases := []testCase{
		{
			name: "basic",
			block: &logspb.BlockLog{
				Id:         bid1,
				GenTraceId: genTrace.Id,

				ExecTraceIds: execTraceIds,
			},
			expected: &logspb.BlockLog{
				Id:             bid1,
				GenTraceId:     genTrace.Id,
				ExecTraceIds:   execTraceIds,
				Doc:            genTrace.GetGenerate().Request.Doc,
				GeneratedBlock: genTrace.GetGenerate().Response.Blocks[0],
				ExecutedBlock:  execTrace2.GetExecute().Request.Block,
				ExitCode:       7,
				EvalMode:       false,
			},
			traces: traces,
		},
		{
			name: "eval_mode",
			block: &logspb.BlockLog{
				Id:         bid2,
				GenTraceId: genTrace2.Id,

				ExecTraceIds: []string{execTrace3.Id},
			},
			expected: &logspb.BlockLog{
				Id:             bid2,
				GenTraceId:     genTrace2.Id,
				ExecTraceIds:   []string{execTrace3.Id},
				Doc:            genTrace2.GetGenerate().Request.Doc,
				GeneratedBlock: genTrace2.GetGenerate().Response.Blocks[0],
				ExecutedBlock:  execTrace3.GetExecute().Request.Block,
				ExitCode:       7,
				EvalMode:       true,
			},
			traces: traces,
		},
	}

	tracesDBDir, err := os.MkdirTemp("", "tracesdb")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	tracesDB, err := pebble.Open(tracesDBDir, &pebble.Options{})
	if err != nil {
		t.Errorf("could not open traces database %s", tracesDBDir)
	}
	defer helpers.DeferIgnoreError(tracesDB.Close)

	for _, trace := range traces {

		if err := dbutil.SetProto(tracesDB, trace.Id, trace); err != nil {
			t.Fatalf("Failed to set trace for key %s: %v", trace.Id, err)
		}
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := buildBlockLog(context.Background(), c.block, tracesDB); err != nil {
				t.Fatalf("buildBlockLog failed: %v", err)
			}

			if d := cmp.Diff(c.expected, c.block, cmpopts.IgnoreUnexported(logspb.BlockLog{}), testutil.BlockComparer, cmpopts.IgnoreUnexported(v1alpha1.Doc{})); d != "" {
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

	oDir, err := os.MkdirTemp("", "analyzeTestDBs")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	tracesDBDir := filepath.Join(oDir, "traces")
	blocksDBDir := filepath.Join(oDir, "blocks")

	if err := a.Analyze(context.Background(), testDir, tracesDBDir, blocksDBDir); err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	t.Logf("Output written to: %s", oDir)

	blocksDB, err := pebble.Open(blocksDBDir, &pebble.Options{})
	if err != nil {
		t.Fatalf("could not open blocks database %s", blocksDBDir)
	}
	defer helpers.DeferIgnoreError(blocksDB.Close)

	actual := map[string]*logspb.BlockLog{}

	iter, err := blocksDB.NewIterWithContext(context.Background(), nil)
	if err != nil {
		t.Fatalf("Failed to create iterator: %v", err)
	}
	defer iter.Close()

	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		if key == nil {
			break
		}

		value, err := iter.ValueAndErr()
		if err != nil {
			t.Fatalf("Failed to read block for key %s; error: %+v", string(key), err)
		}

		b := &logspb.BlockLog{}
		if err := proto.Unmarshal(value, b); err != nil {
			t.Fatalf("Failed to unmarshal block log: %v", err)
		}
		actual[b.Id] = b
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
	block, ok := actual["23706965-8e3b-440d-ba1a-1e1cc035fbd4"]
	if !ok {
		t.Fatalf("Failed to find block with ID: 23706965-8e3b-440d-ba1a-1e1cc035fbd4")
	}
	if block.GenTraceId == "" {
		t.Errorf("Expected GenTraceID to be set")
	}
	if len(block.ExecTraceIds) == 0 {
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
}

func Test_CombineGenerateEntries(t *testing.T) {
	type testCase struct {
		name             string
		linesFile        string
		expectedEvalMode bool
	}

	cases := []testCase{
		{
			name:             "basic",
			linesFile:        "generate_trace_lines.jsonl",
			expectedEvalMode: false,
		},
		{
			name:             "evalMode",
			linesFile:        "generate_trace_lines_eval_mode.jsonl",
			expectedEvalMode: true,
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

			genTrace := trace.GetGenerate()
			if genTrace == nil {
				t.Fatalf("Expected trace to have a generate trace")
			}
			// Assert the trace has a request and a response
			if genTrace.Request == nil {
				t.Errorf("Expected trace to have a request")
			}
			if genTrace.Response == nil {
				t.Errorf("Expected trace to have a response")
			}

			if trace.EvalMode != c.expectedEvalMode {
				t.Errorf("Expected EvalMode to be %v but got %v", c.expectedEvalMode, trace.EvalMode)
			}
		})
	}
}

func Test_CombineExecuteEntries(t *testing.T) {
	type testCase struct {
		name             string
		linesFile        string
		expectedEvalMode bool
	}

	cases := []testCase{
		{
			name:             "basic",
			linesFile:        "execute_traces_lines.jsonl",
			expectedEvalMode: false,
		},
		{
			name:             "eval_mode_true",
			linesFile:        "execute_traces_lines_eval_mode.jsonl",
			expectedEvalMode: true,
		},
		{
			name:             "eval_mode_false",
			linesFile:        "execute_traces_lines_eval_mode_false.jsonl",
			expectedEvalMode: false,
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

			execTrace := trace.GetExecute()
			if execTrace == nil {
				t.Fatalf("Expected trace to have an execute trace")
			}
			// Assert the trace has a request and a response
			if execTrace.Request == nil {
				t.Errorf("Expected trace to have a request")
			}
			if execTrace.Response == nil {
				t.Errorf("Expected trace to have a response")
			}
			if trace.EvalMode != c.expectedEvalMode {
				t.Errorf("Expected EvalMode to be %v but got %v", c.expectedEvalMode, trace.EvalMode)
			}
		})
	}
}

func Test_CombineRunmeEntries(t *testing.T) {
	type testCase struct {
		name             string
		linesFile        string
		expectedEvalMode bool
	}

	cases := []testCase{
		{
			name:             "basic",
			linesFile:        "runme_traces_lines.jsonl",
			expectedEvalMode: false,
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
			trace, err := combineRunMeTrace(context.Background(), entries)
			if err != nil {
				t.Fatalf("combineRunMeTrace failed: %+v", err)
			}
			if trace == nil {
				t.Fatalf("combineRunMeTrace should have returned non nil response")
			}

			rTrace := trace.GetRunMe()
			if rTrace == nil {
				t.Fatalf("Expected trace to have a runme trace")
			}
			// Assert the trace has a request and no response
			if rTrace.Request == nil {
				t.Errorf("Expected trace to have a request")
			}
			// TODO(jeremy): We don't currently log the response with RunMe
			// https://github.com/stateful/runme/blob/6e56cfae38c5a72193a86677356927e14ce87b27/internal/runner/service.go#L461
			if rTrace.Response != nil {
				t.Errorf("Expected trace not to have a response")
			}
			if trace.EvalMode != c.expectedEvalMode {
				t.Errorf("Expected EvalMode to be %v but got %v", c.expectedEvalMode, trace.EvalMode)
			}
		})
	}
}

func Test_updateBlockForExecution(t *testing.T) {
	type testCase struct {
		name     string
		block    *logspb.BlockLog
		trace    *logspb.Trace
		expected *logspb.BlockLog
	}

	cases := []testCase{
		{
			name:  "ExecuteTrace",
			block: &logspb.BlockLog{},
			trace: &logspb.Trace{
				Data: &logspb.Trace_Execute{
					Execute: &logspb.ExecuteTrace{
						Request: &v1alpha1.ExecuteRequest{
							Block: &v1alpha1.Block{
								Contents: "echo hello",
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
					},
				},
			},
			expected: &logspb.BlockLog{
				ExecutedBlock: &v1alpha1.Block{
					Contents: "echo hello",
				},
				ExitCode: 4,
			},
		},
		{
			name:  "RunMeTrace",
			block: &logspb.BlockLog{},
			trace: &logspb.Trace{
				Data: &logspb.Trace_RunMe{
					RunMe: &logspb.RunMeTrace{
						Request: &runnerv1.ExecuteRequest{
							Commands: []string{"prog1", "arg1"},
						},
					},
				},
			},
			expected: &logspb.BlockLog{
				ExecutedBlock: &v1alpha1.Block{
					Contents: "prog1 arg1",
					Kind:     v1alpha1.BlockKind_CODE,
				},
				ExitCode: -2377,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := updateBlockForExecution(c.block, c.trace); err != nil {
				t.Fatalf("updateBlockForExecution failed: %v", err)
			}
			if d := cmp.Diff(c.expected, c.block, cmpopts.IgnoreUnexported(logspb.BlockLog{}), testutil.BlockComparer); d != "" {
				t.Errorf("Unexpected diff:\n%s", d)
			}
		})
	}
}
