package analyze

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jlewi/foyle/app/api"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jlewi/foyle/app/pkg/docs"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/jlewi/monogo/helpers"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	// unsetExitCode is a random negative exit code so we can tell when it hasn't been set.
	unsetExitCode = -2377
)

// Analyzer is responsible for analyzing logs.
type Analyzer struct {
}

// NewAnalyzer creates a new Analyzer.
func NewAnalyzer() (*Analyzer, error) {
	return &Analyzer{}, nil
}

type ResultFiles struct {
	BlockLogs      []string
	GenerateTraces []string
	ExecuteTraces  []string
}

// Analyze analyzes the logs.
func (a *Analyzer) Analyze(ctx context.Context, logsDir string, outDir string) (ResultFiles, error) {
	log := logs.FromContext(ctx)
	log.Info("Analyzing logs", "logsDir", logsDir, "outDir", outDir)

	results := initResultFiles(outDir)

	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		// Logger won't be setup yet so we can't use it.
		log.Info("Creating output directory", "dir", outDir)
		err := os.MkdirAll(outDir, 0755)
		if err != nil {
			return results, errors.Wrapf(err, "could not create log directory %s", outDir)
		}
	}

	jsonFiles, err := findLogFiles(ctx, logsDir)
	if err != nil {
		return results, err
	}

	traces, blocks, err := buildTraces(ctx, jsonFiles, results)
	if err != nil {
		return results, err
	}

	err = buildBlockLogs(ctx, traces, blocks, results.BlockLogs[0])

	return results, err
}

// buildTraces creates a map of all the traces and initializes the blocks.
func buildTraces(ctx context.Context, jsonFiles []string, resultFiles ResultFiles) (map[string]api.Trace, map[string]*api.BlockLog, error) {
	log := logs.FromContext(ctx)
	// Entries is a mapping from a traceId to a list of logEntries associated with that entry.
	traceEntries := make(map[string][]*api.LogEntry)

	for _, p := range jsonFiles {
		log.Info("Reading file", "path", p)
		f, err := os.Open(p)
		if err != nil {
			log.Error(err, "Error opening file; file will be skipped", "path", p)
			continue
		}
		d := json.NewDecoder(f)

		for {
			entry := &api.LogEntry{}
			err := d.Decode(entry)

			if err != nil {
				if err == io.EOF {
					break
				}
				log.Error(err, "Error decoding log entry", "path", p)
				continue
			}

			// Ignore log entries without traces
			if entry.TraceID() == "" {
				continue
			}

			items, ok := traceEntries[entry.TraceID()]
			if !ok {
				items = make([]*api.LogEntry, 0, 10)
			}
			items = append(items, entry)
			traceEntries[entry.TraceID()] = items
		}
	}

	// Store a map of all traces
	traces := make(map[string]api.Trace)

	// Build a map of the blocks
	blocks := make(map[string]*api.BlockLog)

	// Create encoders to write the traces
	genFile, err := os.Create(resultFiles.GenerateTraces[0])
	if err != nil {
		return nil, nil, err
	}
	defer helpers.DeferIgnoreError(genFile.Close)

	execFile, err := os.Create(resultFiles.ExecuteTraces[0])
	if err != nil {
		return nil, nil, err
	}
	defer helpers.DeferIgnoreError(execFile.Close)

	genEnc := json.NewEncoder(genFile)
	execEnc := json.NewEncoder(execFile)

	// Now combine all the entries for each trace
	for tid, items := range traceEntries {
		log.Info("Combining entries for trace", "traceId", tid, "numEntries", len(items))
		trace, err := combineEntriesForTrace(ctx, items)
		if err != nil {
			log.Error(err, "Error combining entries for trace", "traceId", tid)
			continue
		}
		traces[tid] = trace

		var enc *json.Encoder

		// Update the blocks associated with this trace
		switch t := trace.(type) {
		case *api.GenerateTrace:
			for _, oBlock := range t.Response.GetBlocks() {
				bid := oBlock.GetId()
				if bid == "" {
					continue
				}
				block, ok := blocks[bid]
				if !ok {
					block = &api.BlockLog{
						ID: bid,
					}
					blocks[bid] = block
				}
				block.GenTraceID = tid
			}
			enc = genEnc
		case *api.ExecuteTrace:
			bid := t.Request.GetBlock().GetId()
			if bid == "" {
				continue
			}
			block, ok := blocks[bid]
			if !ok {
				block = &api.BlockLog{
					ID: bid,
				}
				blocks[bid] = block
			}
			if block.ExecTraceIDs == nil {
				block.ExecTraceIDs = make([]string, 0, 10)
			}
			block.ExecTraceIDs = append(block.ExecTraceIDs, tid)
			enc = execEnc
		default:
			log.Error(fmt.Errorf("Unknown trace type"), "Unknown trace type", "trace", t)
		}

		if enc != nil {
			if err := enc.Encode(trace); err != nil {
				log.Error(err, "Error writing trace to output file")
			}
		}
	}

	return traces, blocks, nil
}

func findLogFiles(ctx context.Context, logsDir string) ([]string, error) {
	log := logs.FromContext(ctx)
	jsonFiles := []string{}
	paths := map[string]bool{}

	if _, err := os.Stat(logsDir); err != nil && os.IsNotExist(err) {
		return jsonFiles, fmt.Errorf("Analyze invoked for non-existent path: %v", logsDir)
	}

	// Walk the directory and add all JSON files.
	walkErr := filepath.Walk(logsDir,
		func(path string, info os.FileInfo, walkErr error) error {
			// Skip non YAML files
			ext := strings.ToLower(filepath.Ext(info.Name()))

			if ext != ".json" && ext != ".jsonl" {
				return nil
			}
			p, err := filepath.EvalSymlinks(path)
			if err != nil {
				log.Error(err, "Failed to evaluate symlink", "path", path)
				return err
			}
			paths[p] = true
			return nil
		})

	if walkErr != nil {
		return jsonFiles, walkErr
	}

	for p := range paths {
		jsonFiles = append(jsonFiles, p)
	}

	sort.Strings(jsonFiles)

	return jsonFiles, nil
}

func initResultFiles(outDir string) ResultFiles {
	stamp := time.Now().Format("2006-01-02T15:04:05")
	return ResultFiles{
		BlockLogs:      []string{filepath.Join(outDir, fmt.Sprintf("blocks.logs.%s.jsonl", stamp))},
		GenerateTraces: []string{filepath.Join(outDir, fmt.Sprintf("traces.generate.%s.jsonl", stamp))},
		ExecuteTraces:  []string{filepath.Join(outDir, fmt.Sprintf("traces.execute.%s.jsonl", stamp))},
	}
}

func buildBlockLogs(ctx context.Context, traces map[string]api.Trace, blocks map[string]*api.BlockLog, outFile string) error {
	log := logs.FromContext(ctx)

	oDir := filepath.Dir(outFile)
	if _, err := os.Stat(oDir); os.IsNotExist(err) {
		log.Info("Creating directory for block logs", "dir", oDir)
		err := os.MkdirAll(oDir, 0755)
		if err != nil {
			return errors.Wrapf(err, "could not log directory %s", oDir)
		}
	}

	// Now we can process each block and write the combined entries to the output file.
	oFile, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer helpers.DeferIgnoreError(oFile.Close)

	enc := json.NewEncoder(oFile)
	for bid, blockLog := range blocks {
		log.Info("Combining entries for block", "blockId", bid)

		if err := buildBlockLog(ctx, blockLog, traces); err != nil {
			log.Error(err, "Error combining entries for block", "blockId", bid)
			continue
		}
		if err := enc.Encode(blockLog); err != nil {
			log.Error(err, "Error writing example to output file")
		}
	}
	return nil
}

func buildBlockLog(ctx context.Context, block *api.BlockLog, traces map[string]api.Trace) error {
	log := logs.FromContext(ctx)
	log = log.WithValues("blockId", block.ID)
	log.Info("Building block log", "block", block)

	if block.ID == "" {
		return errors.New("Block ID is required")
	}

	if block.GenTraceID != "" {
		genTrace, ok := traces[block.GenTraceID].(*api.GenerateTrace)
		if !ok {
			log.Error(errors.New("Missing GenerateTrace for traceId"), "Error getting generate trace", "genTraceId", block.GenTraceID)
		} else {
			block.Doc = genTrace.Request.GetDoc()
		}

		// Find the actual block
		for _, b := range genTrace.Response.GetBlocks() {
			if b.GetId() == block.ID {
				block.GeneratedBlock = b
				break
			}
		}
		if block.GeneratedBlock == nil {
			log.Error(errors.New("Failed to find generated block"), "Error finding generated block", "blockId", block.ID)
		}
	}

	var lastTrace *api.ExecuteTrace
	// Get the last execution trace
	for _, tid := range block.ExecTraceIDs {
		trace, ok := traces[tid].(*api.ExecuteTrace)
		if !ok {
			log.Error(errors.New("Missing ExecuteTrace for traceId"), "Error getting execute trace", "execTraceId", tid)
			continue
		}

		if lastTrace == nil {
			lastTrace = trace
			continue
		}

		if lastTrace.StartTime.Before(trace.StartTime) {
			lastTrace = trace
		}
	}
	if lastTrace != nil {
		block.ExecutedBlock = lastTrace.Request.GetBlock()
		block.ExitCode = unsetExitCode
		for _, o := range lastTrace.Response.GetOutputs() {
			exitCode, ok := docs.GetExitCode(o)
			if ok {
				block.ExitCode = exitCode
				break
			}
		}
	}

	return nil
}

func combineEntriesForTrace(ctx context.Context, entries []*api.LogEntry) (api.Trace, error) {
	// First sort the entries by timestamp.
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Time().Before(entries[j].Time())
	})

	// Loop through the entries until we identify the message that tells us what kind of trace it is.
	for _, logEntry := range entries {
		function := logEntry.Function()
		if strings.HasSuffix(function, "agent.(*Agent).Generate") {
			return combineGenerateTrace(ctx, entries)
		}

		if strings.HasSuffix(function, "executor.(*Executor).Execute") {
			return combineExecuteTrace(ctx, entries)
		}
	}

	return nil, errors.New("Failed to identify trace type")
}

func combineGenerateTrace(ctx context.Context, entries []*api.LogEntry) (*api.GenerateTrace, error) {
	trace := &api.GenerateTrace{}
	for _, e := range entries {
		if trace.TraceID == "" {
			trace.TraceID = e.TraceID()
		}
		if trace.Request == nil {
			raw := e.Request()
			if raw != nil {
				request := &v1alpha1.GenerateRequest{}
				if err := protojson.Unmarshal([]byte(raw), request); err != nil {
					return nil, err
				}

				trace.Request = request
				trace.StartTime = e.Time()
			}
		}
		if trace.Response == nil {
			raw := e.Response()
			if raw != nil {
				v := &v1alpha1.GenerateResponse{}
				if err := protojson.Unmarshal([]byte(raw), v); err != nil {
					return nil, err
				}
				trace.Response = v
				trace.EndTime = e.Time()
			}
		}
	}

	return trace, nil
}

func combineExecuteTrace(ctx context.Context, entries []*api.LogEntry) (*api.ExecuteTrace, error) {
	trace := &api.ExecuteTrace{}
	for _, e := range entries {
		if trace.TraceID == "" {
			trace.TraceID = e.TraceID()
		}
		if trace.Request == nil {
			raw := e.Request()
			if raw != nil {
				request := &v1alpha1.ExecuteRequest{}
				if err := protojson.Unmarshal([]byte(raw), request); err != nil {
					return nil, err
				}

				trace.Request = request
				trace.StartTime = e.Time()
			}
		}
		if trace.Response == nil {
			raw := e.Response()
			if raw != nil {
				v := &v1alpha1.ExecuteResponse{}
				if err := protojson.Unmarshal([]byte(raw), v); err != nil {
					return nil, err
				}
				trace.Response = v
				trace.EndTime = e.Time()
			}
		}
	}

	return trace, nil
}
