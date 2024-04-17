package analyze

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jlewi/foyle/app/pkg/docs"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/jlewi/monogo/helpers"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

// Analyze analyzes the logs.
func (a *Analyzer) Analyze(ctx context.Context, logsDir string, outFile string) error {
	// Should we support appending to the output file
	oFile, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer helpers.DeferIgnoreError(oFile.Close)

	log := logs.FromContext(ctx)
	log.Info("Analyzing logs", "logsDir", logsDir, "outFile", outFile)
	paths := map[string]bool{}

	if _, err := os.Stat(logsDir); err != nil && os.IsNotExist(err) {
		return fmt.Errorf("Analyze invoked for non-existent path: %v", logsDir)
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
		return walkErr
	}

	jsonFiles := []string{}
	for p := range paths {
		jsonFiles = append(jsonFiles, p)
	}

	sort.Strings(jsonFiles)

	// Entries is a mapping from a traceId to a list of logEntries associated with that entry.
	traceEntries := make(map[string][]*LogEntry)

	for _, p := range jsonFiles {
		log.Info("Reading file", "path", p)
		f, err := os.Open(p)
		if err != nil {
			log.Error(err, "Error opening file; file will be skipped", "path", p)
			continue
		}
		d := json.NewDecoder(f)

		for {
			entry := &LogEntry{}
			err := d.Decode(entry)

			if err != nil {
				if err == io.EOF {
					break
				}
				log.Error(err, "Error decoding log entry")
				continue
			}

			// Ignore log entries without traces
			if entry.TraceID() == "" {
				continue
			}

			items, ok := traceEntries[entry.TraceID()]
			if !ok {
				items = make([]*LogEntry, 0, 10)
			}
			items = append(items, entry)
			traceEntries[entry.TraceID()] = items
		}
	}

	// Store a map of all traces
	traces := make(map[string]Trace)

	// Build a map of the blocks
	blocks := make(map[string]*BlockLog)

	// Now combine all the entries for each trace
	for tid, items := range traceEntries {
		log.Info("Combining entries for trace", "traceId", tid, "numEntries", len(items))
		trace, err := combineEntriesForTrace(ctx, items)
		if err != nil {
			log.Error(err, "Error combining entries for trace", "traceId", tid)
			continue
		}
		traces[tid] = trace

		// Update the blocks associated with this trace
		switch t := trace.(type) {
		case *GenerateTrace:
			for _, oBlock := range t.Response.GetBlocks() {
				bid := oBlock.GetId()
				if bid == "" {
					continue
				}
				block, ok := blocks[bid]
				if !ok {
					block = &BlockLog{
						ID: bid,
					}
					blocks[bid] = block
				}
				block.GenTraceID = tid
			}
		case *ExecuteTrace:
			bid := t.Request.GetBlock().GetId()
			if bid == "" {
				continue
			}
			block, ok := blocks[bid]
			if !ok {
				block = &BlockLog{
					ID: bid,
				}
				blocks[bid] = block
			}
			if block.ExecTraceIDs == nil {
				block.ExecTraceIDs = make([]string, 0, 10)
			}
			block.ExecTraceIDs = append(block.ExecTraceIDs, tid)
		default:
			log.Error(fmt.Errorf("Unknown trace type"), "Unknown trace type", "trace", t)
		}
	}

	// Now we can process each block and write the combined entries to the output file.
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

func buildBlockLog(ctx context.Context, block *BlockLog, traces map[string]Trace) error {
	log := logs.FromContext(ctx)
	log = log.WithValues("blockId", block.ID)
	log.Info("Building block log", "block", block)

	if block.ID == "" {
		return errors.New("Block ID is required")
	}

	if block.GenTraceID != "" {
		genTrace, ok := traces[block.GenTraceID].(*GenerateTrace)
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

	var lastTrace *ExecuteTrace
	// Get the last execution trace
	for _, tid := range block.ExecTraceIDs {
		trace, ok := traces[tid].(*ExecuteTrace)
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
			exitCode, ok := docs.GetExitCode(*o)
			if ok {
				block.ExitCode = exitCode
				break
			}
		}
	}

	return nil
}

func combineEntriesForTrace(ctx context.Context, entries []*LogEntry) (Trace, error) {
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

func combineGenerateTrace(ctx context.Context, entries []*LogEntry) (*GenerateTrace, error) {
	trace := &GenerateTrace{}
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

func combineExecuteTrace(ctx context.Context, entries []*LogEntry) (*ExecuteTrace, error) {
	trace := &ExecuteTrace{}
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
