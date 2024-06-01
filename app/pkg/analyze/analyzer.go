package analyze

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/jlewi/foyle/app/pkg/docs"
	runnerv1 "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/runner/v1"

	"github.com/cockroachdb/pebble"
	"github.com/jlewi/foyle/app/api"
	"github.com/jlewi/foyle/app/pkg/dbutil"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"k8s.io/client-go/util/workqueue"
)

const (
	// unsetExitCode is a random negative exit code so we can tell when it hasn't been set.
	unsetExitCode = -2377
)

// Analyzer is responsible for analyzing logs.
type Analyzer struct {
	tracesDB            *pebble.DB
	blocksDB            *pebble.DB
	queue               workqueue.DelayingInterface
	handleLogFileIsDone sync.WaitGroup
	waterMarks          map[string]int64
}

// NewAnalyzer creates a new Analyzer.
func NewAnalyzer(tracesDB *pebble.DB, blocksDB *pebble.DB) (*Analyzer, error) {
	return &Analyzer{
		tracesDB: tracesDB,
		blocksDB: blocksDB,
		queue:    workqueue.NewDelayingQueue(),
	}, nil
}

type fileItem struct {
	path string
}

// Run runs the analyzer; continually processing logs.
func (a *Analyzer) Run(ctx context.Context, logDirs []string) error {
	// Find all the current files
	jsonFiles, err := findLogFilesInDirs(ctx, logDirs)
	if err != nil {
		return err
	}

	// Enqueue an item to process each file
	for _, f := range jsonFiles {
		a.queue.Add(&fileItem{path: f})
	}

	if err := registerDirWatchers(ctx, a.queue, logDirs); err != nil {
		return err
	}

	a.handleLogFileIsDone.Add(1)

	go a.handleLogFileEvents(ctx)
	return nil
}

// TODO(jeremy): How do we make the Analyzer thread safe? I believe the DB classes are thread safe
func (a *Analyzer) handleLogFileEvents(ctx context.Context) {
	q := a.queue
	log := logs.FromContext(ctx)
	for {
		item, shutdown := q.Get()
		if shutdown {
			a.handleLogFileIsDone.Done()
			return
		}
		func() {
			defer q.Done(item)
			fileItem, ok := item.(*fileItem)
			if !ok {
				log.Error(errors.New("Failed to cast item to fileItem"), "Failed to cast item to fileItem")
				return
			}

			if err := a.processLogFile(ctx, fileItem.path); err != nil {
				log.Error(err, "Error processing log file", "path", fileItem.path)
			}
		}()
	}
}

func (a *Analyzer) processLogFile(ctx context.Context, path string) error {
	return errors.New("Not implemented")
	//log := logs.FromContext(ctx)
	//log.Info("Processing log file", "path", path)
	//
	//f, err := os.Open(path)
	//if err != nil {
	//	return errors.Wrapf(err, "Failed to open file %s", path)
	//}
	//defer f.Close()
	//offset, ok := a.waterMarks[path]
	//if !ok {
	//	offset = 0
	//}
	//startPos, err := f.Seek(offset, 0)
	//if err != nil {
	//	return errors.Wrapf(err, "Failed to seek to offset %d in file %s", offset, path)
	//}
	//
	//scanner := bufio.NewScanner(f)
	//for scanner.Scan() {
	//
	//}
	//
	//if scanner.Err() != nil {
	//	return errors.Wrapf(scanner.Err(), "Error scanning file %s", path)
	//}
	//
	//scanner.Bytes()
	//d := json.NewDecoder(f)
	//
	//for {
	//	entry := &api.LogEntry{}
	//	err := d.Decode(entry)
	//
	//	if err != nil {
	//		if err == io.EOF {
	//			break
	//		}
	//		log.Error(err, "Error decoding log entry", "path", path)
	//		continue
	//	}
	//
	//	// Ignore log entries without traces
	//	if entry.TraceID() == "" {
	//		continue
	//	}
	//
	//	items, ok := traceEntries[entry.TraceID()]
	//	if !ok {
	//		items = make([]*api.LogEntry, 0, 10)
	//	}
	//	items = append(items, entry)
	//	traceEntries[entry.TraceID()] = items
	//}
	//
	//// Now combine all the entries for each trace
	//for tid, items := range traceEntries {
	//	log.Info("Combining entries for trace", "traceId", tid, "numEntries", len(items))
	//	trace, err := combineEntriesForTrace(ctx, items)
	//	if err != nil {
	//		log.Error(err, "Error combining entries for trace", "traceId
	//	}
	//}
}

// registerDirWatchers sets up notifications for changes in the log directories.
// Any time a file is modified it will enqueue the file for processing.
func registerDirWatchers(ctx context.Context, q workqueue.DelayingInterface, logDirs []string) error {
	return errors.New("Not implemented")
	//for _, dir := range logDirs {
	//	return errors.New("Not implemented")
	//}
	//return nil
}

// Analyze analyzes the logs.
// logsDir - Is the directory containing the logs
// tracesDBDir - Is the directory containing the traces pebble database
// blocksDBDir - Is the directory containing the blocks pebble database
//
// TODO(https://github.com/jlewi/foyle/issues/126): I think we need to pass the DBs in because we can't
// have them be multi process; so we probably want to have a single DB per process.
func (a *Analyzer) Analyze(ctx context.Context, logDirs []string) error {
	log := logs.FromContext(ctx)
	log.Info("Analyzing logs", "logDirs", logDirs)

	jsonFiles, err := findLogFilesInDirs(ctx, logDirs)
	if err != nil {
		return err
	}

	if err := buildTraces(ctx, jsonFiles, a.tracesDB, a.blocksDB); err != nil {
		return err
	}

	return buildBlockLogs(ctx, a.tracesDB, a.blocksDB)
}

func (a *Analyzer) Shutdown(ctx context.Context) error {
	log := logs.FromContext(ctx)
	log.Info("Shutting down analyzer")
	a.queue.ShutDown()
	a.handleLogFileIsDone.Wait()

	// TODO(jeremy) Persist the watermarks
	log.Info("Analyzer shutdown")
	return nil
}

// buildTraces creates a map of all the traces and initializes the blocks.
func buildTraces(ctx context.Context, jsonFiles []string, tracesDB *pebble.DB, blocksDB *pebble.DB) error {
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

	// Now combine all the entries for each trace
	for tid, items := range traceEntries {
		log.Info("Combining entries for trace", "traceId", tid, "numEntries", len(items))
		trace, err := combineEntriesForTrace(ctx, items)
		if err != nil {
			log.Error(err, "Error combining entries for trace", "traceId", tid)
			continue
		}

		if err := writeProto(tracesDB, tid, trace); err != nil {
			return err
		}

		// Update the blocks associated with this trace
		switch t := trace.Data.(type) {
		case *logspb.Trace_Generate:
			for _, oBlock := range t.Generate.Response.GetBlocks() {
				bid := oBlock.GetId()
				if bid == "" {
					continue
				}
				if err := readModifyWriteBlock(blocksDB, bid, func(block *logspb.BlockLog) error {
					block.Id = bid
					block.GenTraceId = tid
					return nil
				}); err != nil {
					return errors.Wrapf(err, "Failed to set generate trace on block %s", bid)
				}
			}
		case *logspb.Trace_Execute:
			bid := t.Execute.Request.GetBlock().GetId()
			if bid == "" {
				continue
			}

			if err := readModifyWriteBlock(blocksDB, bid, func(block *logspb.BlockLog) error {
				block.Id = bid
				if block.ExecTraceIds == nil {
					block.ExecTraceIds = make([]string, 0, 10)
				}
				block.ExecTraceIds = append(block.ExecTraceIds, tid)
				return nil
			}); err != nil {
				return errors.Wrapf(err, "Failed to set execute trace on block %s", bid)
			}
		case *logspb.Trace_RunMe:
			bid := t.RunMe.Request.GetKnownId()
			if bid == "" {
				continue
			}

			if err := readModifyWriteBlock(blocksDB, bid, func(block *logspb.BlockLog) error {
				block.Id = bid
				if block.ExecTraceIds == nil {
					block.ExecTraceIds = make([]string, 0, 10)
				}
				block.ExecTraceIds = append(block.ExecTraceIds, tid)
				return nil
			}); err != nil {
				return errors.Wrapf(err, "Failed to set RunMe trace on block %s", bid)
			}
		default:
			log.Error(fmt.Errorf("Unknown trace type"), "Unknown trace type", "trace", t)
		}
	}

	return nil
}

func findLogFilesInDirs(ctx context.Context, logDirs []string) ([]string, error) {
	log := logs.FromContext(ctx)
	jsonFiles := make([]string, 0, 100)
	for _, logsDir := range logDirs {
		newFiles, err := findLogFiles(ctx, logsDir)
		if err != nil {
			return jsonFiles, err
		}
		log.Info("Found logs", "numFiles", len(newFiles), "logsDir", logsDir)
		jsonFiles = append(jsonFiles, newFiles...)
	}
	return jsonFiles, nil
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

func buildBlockLogs(ctx context.Context, tracesDB *pebble.DB, blocksDB *pebble.DB) error {
	log := logs.FromContext(ctx)

	iter, err := blocksDB.NewIterWithContext(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		if key == nil {
			break
		}
		bid := string(key)
		value, err := iter.ValueAndErr()
		if err != nil {
			return errors.Wrapf(err, "Failed to read block for key %s", string(key))
		}

		log.Info("Combining entries for block", "blockId", bid)

		blockLog := &logspb.BlockLog{}
		if err := proto.Unmarshal(value, blockLog); err != nil {
			return errors.Wrapf(err, "Failed to unmarshal block for id %s", bid)
		}
		if err := buildBlockLog(ctx, blockLog, tracesDB); err != nil {
			log.Error(err, "Error combining entries for block", "blockId", bid)
			continue
		}
		bytes, err := proto.Marshal(blockLog)
		if err != nil {
			return errors.Wrapf(err, "Failed to marshal block for id %s", bid)
		}
		log.Info("Writing block", "blockId", bid, "block", blockLog)
		if err := blocksDB.Set([]byte(bid), bytes, pebble.Sync); err != nil {
			log.Error(err, "Error writing block", "blockId", bid)
		}
	}
	return nil
}

func buildBlockLog(ctx context.Context, block *logspb.BlockLog, tracesDB *pebble.DB) error {
	log := logs.FromContext(ctx)
	log = log.WithValues("blockId", block.Id)
	log.Info("Building block log", "block", block)

	if block.Id == "" {
		return errors.WithStack(errors.New("Block ID is required"))
	}

	if block.GenTraceId != "" {
		func() {
			trace := &logspb.Trace{}
			if err := dbutil.GetProto(tracesDB, block.GenTraceId, trace); err != nil {
				log.Error(err, "Error getting generate trace", "genTraceId", block.GenTraceId)
				return
			}
			genTrace, ok := trace.Data.(*logspb.Trace_Generate)
			if !ok {
				log.Error(errors.New("Invalid GenerateTrace for traceId"), "Error getting generate trace", "genTraceId", block.GenTraceId)
				return
			}

			block.Doc = genTrace.Generate.Request.GetDoc()
			// If the block was generated as part of evaluation mode then consider it to be in evaluation mode.
			if trace.EvalMode {
				block.EvalMode = true
			}

			// Find the actual block
			for _, b := range genTrace.Generate.Response.GetBlocks() {
				if b.GetId() == block.GetId() {
					block.GeneratedBlock = b
					return
				}
			}
			if block.GeneratedBlock == nil {
				log.Error(errors.New("Failed to find generated block"), "Error finding generated block", "blockId", block.GetId())
			}
		}()
	}

	var lastTrace *logspb.Trace
	// Get the last execution trace
	for _, tid := range block.GetExecTraceIds() {
		func() {
			trace := &logspb.Trace{}
			if err := dbutil.GetProto(tracesDB, tid, trace); err != nil {
				log.Error(err, "Error getting execute trace", "execTraceId", tid)
				return
			}

			if trace.GetExecute() == nil && trace.GetRunMe() == nil {
				log.Error(errors.New("Invalid execution trace for traceId"), "Error getting execute trace", "execTraceId", tid)
				return
			}
			if lastTrace == nil {
				lastTrace = trace
				return
			}

			if lastTrace.StartTime.AsTime().Before(trace.StartTime.AsTime()) {
				lastTrace = trace
			}
		}()
	}
	if lastTrace != nil {
		if err := updateBlockForExecution(block, lastTrace); err != nil {
			return err
		}
	}

	return nil
}

// updateBlockForExecution updates fields in the block log based the last execution trace of that block
func updateBlockForExecution(block *logspb.BlockLog, lastTrace *logspb.Trace) error {
	// If the block was executed as part of evaluation mode then consider it to be in evaluation mode.
	if lastTrace.EvalMode {
		block.EvalMode = true
	}
	block.ExecutedBlock = nil
	block.ExitCode = unsetExitCode

	switch eTrace := lastTrace.Data.(type) {
	case *logspb.Trace_Execute:
		block.ExecutedBlock = eTrace.Execute.Request.GetBlock()

		for _, o := range eTrace.Execute.Response.GetOutputs() {
			exitCode, ok := docs.GetExitCode(o)
			if ok {
				block.ExitCode = int32(exitCode)
				break
			}
		}
	case *logspb.Trace_RunMe:
		// TODO(jeremy): Is this the right way to turn the command into a string?
		block.ExecutedBlock = &v1alpha1.Block{
			Kind:     v1alpha1.BlockKind_CODE,
			Contents: strings.Join(eTrace.RunMe.Request.GetCommands(), " "),
			Outputs:  nil,
		}

	default:
		return errors.WithStack(errors.Errorf("Can't update BlockLog with execution information. The last trace, id %s  is not an execution trace", lastTrace.Id))
	}
	return nil
}

func combineEntriesForTrace(ctx context.Context, entries []*api.LogEntry) (*logspb.Trace, error) {
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

		if strings.HasSuffix(function, "runner.(*runnerService).Execute") {
			return combineRunMeTrace(ctx, entries)
		}
	}

	return nil, errors.New("Failed to identify trace type")
}

func combineGenerateTrace(ctx context.Context, entries []*api.LogEntry) (*logspb.Trace, error) {
	gTrace := &logspb.GenerateTrace{}
	trace := &logspb.Trace{
		Data: &logspb.Trace_Generate{
			Generate: gTrace,
		},
		Spans: make([]*logspb.Span, 0, 10),
	}
	evalMode := false
	for _, e := range entries {
		if trace.Id == "" {
			trace.Id = e.TraceID()
		}
		if mode, present := e.EvalMode(); present {
			// If any of the entries are marked as true then we will consider the trace to be in eval mode.
			// We don't want to assume that the evalMode will be set on all log entries in the trace.
			// So the logic is to assume its not eval mode by default and then set it to eval mode if we find
			// One entry that is marked as eval mode.
			if mode {
				evalMode = mode
			}
		}

		if gTrace.Request == nil {
			raw := e.Request()
			if raw != nil {
				request := &v1alpha1.GenerateRequest{}
				if err := protojson.Unmarshal([]byte(raw), request); err != nil {
					return nil, err
				}

				gTrace.Request = request
				trace.StartTime = timestamppb.New(e.Time())
			}
		}
		if gTrace.Response == nil {
			raw := e.Response()
			if raw != nil {
				v := &v1alpha1.GenerateResponse{}
				if err := protojson.Unmarshal([]byte(raw), v); err != nil {
					return nil, err
				}
				gTrace.Response = v
				trace.EndTime = timestamppb.New(e.Time())
			}
		}

		span := logEntryToSpan(ctx, e)
		if span != nil {
			trace.Spans = append(trace.Spans, span)
		}
	}
	trace.EvalMode = evalMode

	combineSpans(trace)
	return trace, nil
}

func combineExecuteTrace(ctx context.Context, entries []*api.LogEntry) (*logspb.Trace, error) {
	eTrace := &logspb.ExecuteTrace{}
	trace := &logspb.Trace{
		Data: &logspb.Trace_Execute{
			Execute: eTrace,
		},
	}
	evalMode := false
	for _, e := range entries {
		if trace.Id == "" {
			trace.Id = e.TraceID()
		}
		if mode, present := e.EvalMode(); present {
			// If any of the entries are marked as true then we will consider the trace to be in eval mode.
			// We don't want to assume that the evalMode will be set on all log entries in the trace.
			// So the logic is to assume its not eval mode by default and then set it to eval mode if we find
			// One entry that is marked as eval mode.
			if mode {
				evalMode = mode
			}
		}

		if eTrace.Request == nil {
			raw := e.Request()
			if raw != nil {
				request := &v1alpha1.ExecuteRequest{}
				if err := protojson.Unmarshal([]byte(raw), request); err != nil {
					return nil, err
				}

				eTrace.Request = request
				trace.StartTime = timestamppb.New(e.Time())
			}
		}
		if eTrace.Response == nil {
			raw := e.Response()
			if raw != nil {
				v := &v1alpha1.ExecuteResponse{}
				if err := protojson.Unmarshal([]byte(raw), v); err != nil {
					return nil, err
				}
				eTrace.Response = v
				trace.EndTime = timestamppb.New(e.Time())
			}
		}
	}
	trace.EvalMode = evalMode
	return trace, nil
}

func combineRunMeTrace(ctx context.Context, entries []*api.LogEntry) (*logspb.Trace, error) {
	rTrace := &logspb.RunMeTrace{}
	trace := &logspb.Trace{
		Data: &logspb.Trace_RunMe{
			RunMe: rTrace,
		},
	}
	evalMode := false
	for _, e := range entries {
		if trace.Id == "" {
			trace.Id = e.TraceID()
		}
		if mode, present := e.EvalMode(); present {
			// If any of the entries are marked as true then we will consider the trace to be in eval mode.
			// We don't want to assume that the evalMode will be set on all log entries in the trace.
			// So the logic is to assume its not eval mode by default and then set it to eval mode if we find
			// One entry that is marked as eval mode.
			if mode {
				evalMode = mode
			}
		}

		if rTrace.Request == nil {
			raw := e.Request()
			if raw != nil {
				request := &runnerv1.ExecuteRequest{}
				if err := protojson.Unmarshal([]byte(raw), request); err != nil {
					return nil, err
				}

				rTrace.Request = request
				trace.StartTime = timestamppb.New(e.Time())
			}
		}
		if rTrace.Response == nil {
			raw := e.Response()
			if raw != nil {
				v := &runnerv1.ExecuteResponse{}
				if err := protojson.Unmarshal([]byte(raw), v); err != nil {
					return nil, err
				}
				rTrace.Response = v
				trace.EndTime = timestamppb.New(e.Time())
			}
		}
	}
	trace.EvalMode = evalMode
	return trace, nil
}

func writeProto(db *pebble.DB, key string, pb proto.Message) error {
	b, err := proto.Marshal(pb)
	if err != nil {
		return errors.Wrapf(err, "Failed to marshal proto with key %s", key)
	}
	return db.Set([]byte(key), b, pebble.Sync)
}

// readModifyWriteBlock reads a block from the database, modifies it and writes it back.
// If the block doesn't exist an empty BlockLog will be passed to the function.
func readModifyWriteBlock(db *pebble.DB, key string, modify func(*logspb.BlockLog) error) error {
	b, closer, err := db.Get([]byte(key))
	if err != nil && !errors.Is(err, pebble.ErrNotFound) {
		return errors.Wrapf(err, "Failed to read block with key %s", key)
	}
	// Closer is nil on not found
	if closer != nil {
		defer closer.Close()
	}

	block := &logspb.BlockLog{}

	if err != pebble.ErrNotFound {

		if err := proto.Unmarshal(b, block); err != nil {
			return errors.Wrapf(err, "Failed to unmarshal block with key %s", key)
		}
	}

	if err := modify(block); err != nil {
		return errors.Wrapf(err, "Failed to modify block with key %s", key)
	}

	return writeProto(db, key, block)
}
