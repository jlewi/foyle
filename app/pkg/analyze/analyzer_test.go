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

	oDir, err := os.MkdirTemp("", "analyzeTestDBs")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	rawDir := filepath.Join(oDir, "logs")
	rawLogsDBDir := filepath.Join(oDir, "rawlogs")
	tracesDBDir := filepath.Join(oDir, "traces")
	blocksDBDir := filepath.Join(oDir, "blocks")

	rawDB, err := pebble.Open(rawLogsDBDir, &pebble.Options{})
	if err != nil {
		t.Fatalf("could not open blocks database %s", blocksDBDir)
	}
	defer helpers.DeferIgnoreError(rawDB.Close)

	lockingRawDB := NewLockingEntriesDB(rawDB)

	blocksDB, err := pebble.Open(blocksDBDir, &pebble.Options{})
	if err != nil {
		t.Fatalf("could not open blocks database %s", blocksDBDir)
	}
	defer helpers.DeferIgnoreError(blocksDB.Close)

	lockingBlocksDB := NewLockingBlocksDB(blocksDB)

	tracesDB, err := pebble.Open(tracesDBDir, &pebble.Options{})
	if err != nil {
		t.Fatalf("could not open blocks database %s", blocksDBDir)
	}
	defer helpers.DeferIgnoreError(tracesDB.Close)

	if err := os.MkdirAll(rawDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Copy the logs to the raw logs directory
	logContents, err := os.ReadFile(filepath.Join(testDir, "foyle.logs.2024-04-16T19:06:47.json"))
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logFile := filepath.Join(rawDir, "log.json")
	if err := os.WriteFile(logFile, logContents, 0644); err != nil {
		t.Fatalf("Failed to write log file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rawDir, "log.json"), logContents, 0644); err != nil {
		t.Fatalf("Failed to write log file: %v", err)
	}

	logOffsetsFile := filepath.Join(rawDir, "log_offsets.json")
	a, err := NewAnalyzer(logOffsetsFile, lockingRawDB, tracesDB, lockingBlocksDB)
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	// Create a channel for the analyzer to signal when a file has been processed
	fileProcessed := make(chan string, 10)
	blockProccessed := make(chan string, 10)

	a.signalFileDone = fileProcessed
	a.signalBlockDone = blockProccessed

	if err := a.Run(context.Background(), []string{rawDir}); err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	fileDone := <-fileProcessed
	t.Logf("File processed: %s", fileDone)
	t.Logf("Output written to: %s", oDir)

	waitForBlock(t, "23706965-8e3b-440d-ba1a-1e1cc035fbd4", 2, blockProccessed)

	// This is a block that was generated via the AI and then executed so run some additional checks
	block := &logspb.BlockLog{}
	if err := dbutil.GetProto(blocksDB, "23706965-8e3b-440d-ba1a-1e1cc035fbd4", block); err != nil {
		t.Fatalf("Failed to find block with ID: 23706965-8e3b-440d-ba1a-1e1cc035fbd4; error %+v", err)
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

	// Now append some logs to the logFile and see that they get processed
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	if _, err := f.Write([]byte(`{"severity":"info","time":1713319614.911959,"caller":"agent/agent.go:61","function":"github.com/jlewi/foyle/app/pkg/agent.(*Agent).Generate","message":"Agent.Generate","traceId":"newtrace1234","request":{"doc":{"blocks":[{"contents":"echo hello"}]}}}` + "\n")); err != nil {
		t.Fatalf("Failed to write to log file: %v", err)
	}
	if _, err := f.Write([]byte(`{"severity":"info","time":1713319616.654191,"caller":"agent/agent.go:83","function":"github.com/jlewi/foyle/app/pkg/agent.(*Agent).Generate","message":"Agent.Generate returning response","traceId":"newtrace1234","response":{"blocks":[{"kind":"MARKUP","language":"","contents":"To find the merge point","outputs":[],"trace_ids":[],"id":"newblock"}]}}` + "\n")); err != nil {
		t.Fatalf("Failed to write to log file: %v", err)
	}

	// N.B. When I didn't call sync the contents of the file didn't get updated. I would have thought calling
	// close was enough.
	if err := f.Sync(); err != nil {
		t.Fatalf("Failed to sync log file: %v", err)

	}
	if err := f.Close(); err != nil {
		t.Fatalf("Failed to close log file: %v", err)
	}

	// Wait for the new block to be processed
	// N.B. I don't know that we are guaranteed that when waitForBlock returns that we have processed all the log lines
	// for this newTrace. We might have just processed the first line. This should be ok as long as our assertions
	// don't require all log entries
	newBlockID := "newblock"
	waitForBlock(t, newBlockID, 1, blockProccessed)

	newBock := &logspb.BlockLog{}
	if err := dbutil.GetProto(blocksDB, newBlockID, newBock); err != nil {
		t.Fatalf("Failed to find block with ID: %s; error %+v", newBlockID, err)
	}
	if newBock.GenTraceId == "" {
		t.Errorf("Expected GenTraceID to be set")
	}
	if newBock.Doc == nil {
		t.Errorf("Expected Doc to be set")
	}
	if newBock.GeneratedBlock == nil {
		t.Errorf("Expected GeneratedBlock to be set")
	}

	if err := a.Shutdown(context.Background()); err != nil {
		t.Fatalf("Failed to shutdown analyzer: %v", err)
	}
}

func waitForBlock(t *testing.T, blockId string, numExpected int, blockProccessed <-chan string) {
	// This is kludgy and brittle way to to wait for the block to be processed.
	// The blockLog should be triggered N times where N is the number of traces attached to the blocklog.
	// It will be triggered twice; once for the generate trace and once for the execute trace.
	timeout := time.Now().Add(1 * time.Minute)

	numActual := 0
	func() {
		for {
			if time.Now().After(timeout) {
				t.Errorf("Timed out waiting for block to be processed %d times", numExpected)
			}
			blockDone := ""
			select {
			case <-time.After(time.Minute):
				t.Errorf("Timed out waiting for block to be processed")
				return
			case blockDone = <-blockProccessed:
			}

			t.Logf("Block processed: %s", blockDone)
			if blockDone == blockId {
				numActual += 1
				if numActual >= numExpected {
					break
				}
			}
		}
	}()
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
