package analyze

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/app/pkg/logs"

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

	traces[genTrace.Id] = genTrace
	traces[genTrace2.Id] = genTrace2

	cases := []testCase{
		{
			name: "basic",
			block: &logspb.BlockLog{
				Id:         bid1,
				GenTraceId: genTrace.Id,
			},
			expected: &logspb.BlockLog{
				Id:             bid1,
				GenTraceId:     genTrace.Id,
				Doc:            genTrace.GetGenerate().Request.Doc,
				GeneratedBlock: genTrace.GetGenerate().Response.Blocks[0],
				EvalMode:       false,
			},
			traces: traces,
		},
		{
			name: "eval_mode",
			block: &logspb.BlockLog{
				Id:         bid2,
				GenTraceId: genTrace2.Id,
			},
			expected: &logspb.BlockLog{
				Id:             bid2,
				GenTraceId:     genTrace2.Id,
				Doc:            genTrace2.GetGenerate().Request.Doc,
				GeneratedBlock: genTrace2.GetGenerate().Response.Blocks[0],
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

type fakeNotifier struct {
	counts map[string]int
}

func (f *fakeNotifier) PostSession(session *logspb.Session) error {
	if f.counts == nil {
		f.counts = make(map[string]int)
	}
	if _, ok := f.counts[session.GetContextId()]; !ok {
		f.counts[session.GetContextId()] = 0

	}
	f.counts[session.GetContextId()] += 1
	return nil
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

	sessionsPath := filepath.Join(oDir, "sessions.sqllite3")
	db, err := sql.Open(SQLLiteDriver, sessionsPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v; error: %+v", sessionsPath, err)
	}

	sessionsManager, err := NewSessionsManager(db)
	if err != nil {
		t.Fatalf("Failed to create sessions manager: %v", err)
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
	a, err := NewAnalyzer(logOffsetsFile, 3*time.Second, lockingRawDB, tracesDB, lockingBlocksDB, sessionsManager)
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	// Create a channel for the analyzer to signal when a file has been processed
	fileProcessed := make(chan string, 10)
	blockProccessed := make(chan string, 10)

	a.signalFileDone = fileProcessed
	a.signalBlockDone = blockProccessed

	fakeNotifier := &fakeNotifier{}
	if err := a.Run(context.Background(), []string{rawDir}, fakeNotifier.PostSession); err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	fileDone := <-fileProcessed
	t.Logf("File processed: %s", fileDone)
	t.Logf("Output written to: %s", oDir)

	// Wait for the logs to be fully processed
	done := false
	timeend := time.Now().Add(1 * time.Minute)
	var w *logspb.LogsWaterMark
	for !done && time.Now().Before(timeend) {
		w = a.GetWatermark()
		if w.Offset < 23457 {
			time.Sleep(5 * time.Second)
		} else {
			done = true
		}
	}
	if !done {
		t.Fatalf("Timed out waiting for logs to be processed; final offset %d", w.Offset)
	}

	// Signal should be triggered  once for the blocklog.
	expectedBlockID := "23706965-8e3b-440d-ba1a-1e1cc035fbd4"
	waitForBlock(t, expectedBlockID, 1, blockProccessed)

	// This is a block that was generated via the AI and then executed so run some additional checks
	block := &logspb.BlockLog{}
	if err := dbutil.GetProto(blocksDB, expectedBlockID, block); err != nil {
		t.Fatalf("Failed to find block with ID: %s; error %+v", expectedBlockID, err)
	}
	if block.GenTraceId == "" {
		t.Errorf("Expected GenTraceID to be set")
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

func assertHasAssertion(t *logspb.Trace) string {
	if len(t.Assertions) == 0 {
		return "Expected trace to have at least 1 assertion"
	}
	return ""
}

type assertTrace func(t *logspb.Trace) string

func Test_CombineGenerateEntries(t *testing.T) {
	type testCase struct {
		name      string
		linesFile string
		// Optional function to generate some logs to
		logFunc          func(log logr.Logger)
		expectedEvalMode bool

		assertions []assertTrace
	}

	cases := []testCase{
		{
			name:             "basic",
			linesFile:        "generate_trace_lines.jsonl",
			expectedEvalMode: false,
			logFunc: func(log logr.Logger) {
				assertion := &v1alpha1.Assertion{
					Name:   v1alpha1.Assertion_ONE_CODE_CELL,
					Result: v1alpha1.AssertResult_PASSED,
					Detail: "",
					Id:     "1234",
				}
				log.Info(logs.Level1Assertion, "assertion", assertion)
			},
			assertions: []assertTrace{
				assertHasAssertion,
			},
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

			logFiles := []string{
				filepath.Join(cwd, "test_data", c.linesFile),
			}

			if c.logFunc != nil {
				// Create a logger to write the logs to a file
				f, err := os.CreateTemp("", "testlogs.jsonl")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				logFile := f.Name()
				if err := f.Close(); err != nil {
					t.Fatalf("Failed to close file: %v", err)
				}

				t.Log("Log file:", logFile)

				config := zap.NewProductionConfig()
				// N.B. This needs to be kept in sync with the fields set in app.go otherwise our test won't use
				// the same fields as in production.
				config.OutputPaths = []string{f.Name()}
				config.EncoderConfig.LevelKey = "severity"
				config.EncoderConfig.TimeKey = "time"
				config.EncoderConfig.MessageKey = "message"
				// We attach the function key to the logs because that is useful for identifying the function that generated the log.
				config.EncoderConfig.FunctionKey = "function"

				logFiles = append(logFiles, logFile)

				testLog, err := config.Build()
				if err != nil {
					t.Fatalf("Failed to create logger: %v", err)
				}

				zTestLog := zapr.NewLogger(testLog)

				c.logFunc(zTestLog)

				if err := testLog.Sync(); err != nil {
					t.Fatalf("Failed to sync log: %v", err)
				}
			}

			for _, logFile := range logFiles {
				testFile, err := os.Open(logFile)
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

			for _, assert := range c.assertions {
				if msg := assert(trace); msg != "" {
					t.Errorf(msg)
				}
			}
		})
	}
}

func Test_DedupeAssertions(t *testing.T) {
	type testCase struct {
		name     string
		input    []*v1alpha1.Assertion
		expected []*v1alpha1.Assertion
	}

	cases := []testCase{
		{
			name: "basic",
			input: []*v1alpha1.Assertion{
				{
					Name:   v1alpha1.Assertion_ONE_CODE_CELL,
					Result: v1alpha1.AssertResult_PASSED,
					Id:     "1",
				},
				{
					Name:   v1alpha1.Assertion_ONE_CODE_CELL,
					Result: v1alpha1.AssertResult_PASSED,
					Id:     "1",
				},
				{
					Name:   v1alpha1.Assertion_ENDS_WITH_CODE_CELL,
					Result: v1alpha1.AssertResult_PASSED,
					Id:     "2",
				},
			},
			expected: []*v1alpha1.Assertion{
				{
					Name:   v1alpha1.Assertion_ONE_CODE_CELL,
					Result: v1alpha1.AssertResult_PASSED,
					Id:     "1",
				},
				{
					Name:   v1alpha1.Assertion_ENDS_WITH_CODE_CELL,
					Result: v1alpha1.AssertResult_PASSED,
					Id:     "2",
				},
			},
		},
		{
			name:     "nil",
			input:    nil,
			expected: []*v1alpha1.Assertion{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			trace := &logspb.Trace{
				Assertions: c.input,
			}

			dedupeAssertions(trace)

			if d := cmp.Diff(c.expected, trace.Assertions, cmpopts.IgnoreUnexported(v1alpha1.Assertion{})); d != "" {
				t.Errorf("Unexpected diff:\n%s", d)
			}
		})
	}
}
