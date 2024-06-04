package analyze

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fsnotify/fsnotify"
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

// Analyzer is responsible for analyzing logs and building traces. It does this in a streaming fashion so that
// traces get built in "realtime".
//
// The Analyzer is multi-threaded. One potential pitfall is we have multiple writers trying to update the same
// key. This would result in a last-write win situation with the last write potentially overwriting the changes
// by the other writer. To avoid this, we use WorkQueue's to schedule updates to each key. The workqueue
// ensures that a single worker is processing a single key at a time.
type Analyzer struct {
	tracesDB  *pebble.DB
	blocksDB  *pebble.DB
	rawLogsDB *pebble.DB
	// queue for log file processing
	queue workqueue.DelayingInterface
	// Queue for block log processing
	// TODO(jeremy): We should really use a durable queue backed by files
	blockQueue workqueue.DelayingInterface

	watcher *fsnotify.Watcher

	handleLogFileIsDone sync.WaitGroup
	handleBlocksIsDone  sync.WaitGroup
	logFileOffsets      map[string]int64
	mu                  sync.Mutex

	// Only used during testing to allow the test to tell when the log file processing is done.
	signalFileDone  chan<- string
	signalBlockDone chan<- string
}

// NewAnalyzer creates a new Analyzer.
func NewAnalyzer(rawLogsDB *pebble.DB, tracesDB *pebble.DB, blocksDB *pebble.DB) (*Analyzer, error) {
	return &Analyzer{
		rawLogsDB:      rawLogsDB,
		tracesDB:       tracesDB,
		blocksDB:       blocksDB,
		queue:          workqueue.NewDelayingQueue(),
		blockQueue:     workqueue.NewDelayingQueue(),
		logFileOffsets: make(map[string]int64),
	}, nil
}

type fileItem struct {
	path string
}

type blockItem struct {
	id string
	// genTraceId is the traceId of the generate trace that generated this block.
	genTraceId string
	// execTraceIds is the traceIds of the execution traces that executed this block.
	execTraceIds []string
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

	if err := a.registerDirWatchers(ctx, a.queue, logDirs); err != nil {
		return err
	}

	a.handleLogFileIsDone.Add(1)
	a.handleBlocksIsDone.Add(1)

	go a.handleLogFileEvents(ctx)
	go a.handleBlockEvents(ctx)

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
			if a.signalFileDone != nil {
				a.signalFileDone <- fileItem.path
			}
		}()
	}
}

func (a *Analyzer) processLogFile(ctx context.Context, path string) error {
	log := logs.FromContext(ctx)
	log.Info("Processing log file", "path", path)

	offset := a.getLogFileOffset(path)
	lines, offset, err := readLinesFromOffset(ctx, path, offset)
	if err != nil {
		return err
	}

	traceIDs := make(map[string]bool)

	for _, line := range lines {
		entry := &api.LogEntry{}
		if err := json.Unmarshal([]byte(line), entry); err != nil {
			log.Error(err, "Error decoding log entry", "path", path, "line", line)
			continue
		}

		// Ignore log entries without traces
		if entry.TraceID() == "" {
			continue
		}

		entries := &logspb.LogEntries{}
		dbutil.ReadModifyWrite[*logspb.LogEntries](a.rawLogsDB, entry.TraceID(), entries, func(entries *logspb.LogEntries) error {
			if entries.Lines == nil {
				entries.Lines = make([]string, 0, 1)
			}
			entries.Lines = append(entries.Lines, line)
			return nil
		})

		traceIDs[entry.TraceID()] = true
	}

	// Combine the entries for each trace that we saw.
	// N.B. We could potentially make this more efficient by checking if the log message is the final message
	// in a trace. This would avoid potentially doing a combine for a trace on each log message.
	for tid := range traceIDs {
		if err := a.buildTrace(ctx, tid); err != nil {
			log.Error(err, "Error building trace", "traceId", tid)
		}
	}
	// Update the offset
	a.setLogFileOffset(path, offset)
	return nil
}

func (a *Analyzer) getLogFileOffset(path string) int64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	offset, ok := a.logFileOffsets[path]
	if !ok {
		return 0
	}
	return offset
}

func (a *Analyzer) setLogFileOffset(path string, offset int64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.logFileOffsets[path] = offset
}

// registerDirWatchers sets up notifications for changes in the log directories.
// Any time a file is modified it will enqueue the file for processing.
func (a *Analyzer) registerDirWatchers(ctx context.Context, q workqueue.DelayingInterface, logDirs []string) error {
	log := logs.FromContext(ctx)
	watcher, err := fsnotify.NewWatcher()
	a.watcher = watcher
	if err != nil {
		return err
	}
	for _, dir := range logDirs {
		fullPath, err := filepath.Abs(dir)
		if err != nil {
			return errors.Wrapf(err, "Failed to get absolute path for %s", dir)
		}

		log.Info("Watching logs directory", "dir", fullPath)
		if err := watcher.Add(fullPath); err != nil {
			return err
		}
	}

	go handleFsnotifications(ctx, watcher, q)
	return nil
}

// handleFsnotifications processes file system notifications by enqueuing the file for processing.
func handleFsnotifications(ctx context.Context, watcher *fsnotify.Watcher, q workqueue.DelayingInterface) {
	log := logs.FromContext(ctx)
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Info("File modified", "path", event.Name)
				q.Add(&fileItem{path: event.Name})
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Error(err, "Error from watcher")
		}
	}
}

// Analyze analyzes the logs.
// logsDir - Is the directory containing the logs
// tracesDBDir - Is the directory containing the traces pebble database
// blocksDBDir - Is the directory containing the blocks pebble database
//
// TODO(https://github.com/jlewi/foyle/issues/126): I think we need to pass the DBs in because we can't
// have them be multi process; so we probably want to have a single DB per process.
//func (a *Analyzer) Analyze(ctx context.Context, logDirs []string) error {
//	log := logs.FromContext(ctx)
//	log.Info("Analyzing logs", "logDirs", logDirs)
//
//	jsonFiles, err := findLogFilesInDirs(ctx, logDirs)
//	if err != nil {
//		return err
//	}
//
//	if err := buildTraces(ctx, jsonFiles, a.tracesDB, a.blocksDB); err != nil {
//		return err
//	}
//
//	return buildBlockLogs(ctx, a.tracesDB, a.blocksDB)
//}

func (a *Analyzer) Shutdown(ctx context.Context) error {
	log := logs.FromContext(ctx)
	log.Info("Shutting down analyzer")

	// Shutdown the watcher
	if a.watcher != nil {
		a.watcher.Close()
	}
	// Shutdown the queues
	a.queue.ShutDown()
	a.blockQueue.ShutDown()
	// Wait for the queues to be shutdown
	a.handleLogFileIsDone.Wait()
	a.handleBlocksIsDone.Wait()

	// TODO(jeremy) Persist the watermarks
	log.Info("Analyzer shutdown")
	return nil
}

// buildTrace creates the trace and initializes the blocks.
func (a *Analyzer) buildTrace(ctx context.Context, tid string) error {
	log := logs.FromContext(ctx)

	// Entries is a mapping from a traceId to a list of logEntries associated with that entry.
	logEntries := make([]*api.LogEntry, 0, 10)

	protoEntries := &logspb.LogEntries{}
	if err := dbutil.GetProto(a.rawLogsDB, tid, protoEntries); err != nil {
		return err
	}

	for _, l := range protoEntries.Lines {
		entry := &api.LogEntry{}
		if err := json.Unmarshal([]byte(l), entry); err != nil {
			log.Error(err, "Error decoding log entry", "line", l)
			continue
		}
		logEntries = append(logEntries, entry)
	}

	// Now combine the entries for the trace
	log.Info("Combining entries for trace", "traceId", tid, "numEntries", len(logEntries))
	trace, err := combineEntriesForTrace(ctx, logEntries)
	if err != nil {
		log.Error(err, "Error combining entries for trace", "traceId", tid)
		return err
	}

	if err := writeProto(a.tracesDB, tid, trace); err != nil {
		return err
	}

	// Update the blocks associated with this trace
	blockDeltas, err := func() (map[string]blockItem, error) {
		bids := map[string]blockItem{}
		switch t := trace.Data.(type) {
		case *logspb.Trace_Generate:
			for _, oBlock := range t.Generate.Response.GetBlocks() {
				bid := oBlock.GetId()
				if bid == "" {
					continue
				}
				delta, ok := bids[bid]
				if !ok {
					delta = blockItem{}
				}
				delta.id = bid
				delta.genTraceId = tid
				bids[bid] = delta
			}
		case *logspb.Trace_Execute:
			bid := t.Execute.Request.GetBlock().GetId()
			if bid == "" {
				return bids, nil
			}
			delta, ok := bids[bid]
			if !ok {
				delta = blockItem{}
			}
			delta.id = bid
			if delta.execTraceIds == nil {
				delta.execTraceIds = make([]string, 0, 11)
			}
			delta.execTraceIds = append(delta.execTraceIds, tid)
			bids[bid] = delta
		case *logspb.Trace_RunMe:
			bid := t.RunMe.Request.GetKnownId()
			if bid == "" {
				return bids, nil
			}
			delta, ok := bids[bid]
			if !ok {
				delta = blockItem{}
				delta.id = bid
			}
			if delta.execTraceIds == nil {
				delta.execTraceIds = make([]string, 0, 11)
			}
			delta.execTraceIds = append(delta.execTraceIds, tid)
			bids[bid] = delta
		default:
			log.Error(fmt.Errorf("Unknown trace type"), "Unknown trace type", "trace", t)
		}
		return bids, nil
	}()

	// TODO(jeremy): Should we enqueue block update events even if there is an error
	if err != nil {
		return err
	}

	// Enqueue the block updates
	for _, delta := range blockDeltas {
		a.blockQueue.Add(&delta)
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

func (a *Analyzer) handleBlockEvents(ctx context.Context) {
	log := logs.FromContext(ctx)
	for {
		item, shutdown := a.blockQueue.Get()
		if shutdown {
			a.handleBlocksIsDone.Done()
			return
		}
		func() {
			defer a.blockQueue.Done(item)
			blockItem, ok := item.(*blockItem)
			if !ok {
				log.Error(errors.New("Failed to cast item to blockItem"), "Failed to cast item to blockItem")
				return
			}

			err := dbutil.ReadModifyWrite[*logspb.BlockLog](a.blocksDB, blockItem.id, &logspb.BlockLog{}, func(block *logspb.BlockLog) error {
				block.Id = blockItem.id
				if blockItem.genTraceId != "" {
					block.GenTraceId = blockItem.genTraceId
				}
				if blockItem.execTraceIds != nil {
					if block.ExecTraceIds == nil {
						block.ExecTraceIds = make([]string, 0, len(blockItem.execTraceIds))
					}
					block.ExecTraceIds = append(block.ExecTraceIds, blockItem.execTraceIds...)
				}
				return buildBlockLog(ctx, block, a.tracesDB)
			})
			if err != nil {
				log.Error(err, "Error processing block", "block", blockItem.id)
			}
			if a.signalBlockDone != nil {
				a.signalBlockDone <- blockItem.id
			}
		}()
	}
}

//func buildBlockLogs(ctx context.Context, tracesDB *pebble.DB, blocksDB *pebble.DB) error {
//	log := logs.FromContext(ctx)
//
//	iter, err := blocksDB.NewIterWithContext(ctx, nil)
//	if err != nil {
//		return err
//	}
//	defer iter.Close()
//
//	for iter.First(); iter.Valid(); iter.Next() {
//		key := iter.Key()
//		if key == nil {
//			break
//		}
//		bid := string(key)
//		value, err := iter.ValueAndErr()
//		if err != nil {
//			return errors.Wrapf(err, "Failed to read block for key %s", string(key))
//		}
//
//		log.Info("Combining entries for block", "blockId", bid)
//
//		blockLog := &logspb.BlockLog{}
//		if err := proto.Unmarshal(value, blockLog); err != nil {
//			return errors.Wrapf(err, "Failed to unmarshal block for id %s", bid)
//		}
//		if err := buildBlockLog(ctx, blockLog, tracesDB); err != nil {
//			log.Error(err, "Error combining entries for block", "blockId", bid)
//			continue
//		}
//		bytes, err := proto.Marshal(blockLog)
//		if err != nil {
//			return errors.Wrapf(err, "Failed to marshal block for id %s", bid)
//		}
//		log.Info("Writing block", "blockId", bid, "block", blockLog)
//		if err := blocksDB.Set([]byte(bid), bytes, pebble.Sync); err != nil {
//			log.Error(err, "Error writing block", "blockId", bid)
//		}
//	}
//	return nil
//}

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

	// Dedupe the execution traces just in case
	uEids := make(map[string]bool)
	for _, eid := range block.GetExecTraceIds() {
		uEids[eid] = true
	}
	block.ExecTraceIds = make([]string, 0, len(uEids))
	for eid := range uEids {
		block.ExecTraceIds = append(block.ExecTraceIds, eid)
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

// TODO(jeremy): We should be able to replace this with the code in dbutil
func writeProto(db *pebble.DB, key string, pb proto.Message) error {
	b, err := proto.Marshal(pb)
	if err != nil {
		return errors.Wrapf(err, "Failed to marshal proto with key %s", key)
	}
	return db.Set([]byte(key), b, pebble.Sync)
}

// readModifyWriteBlock reads a block from the database, modifies it and writes it back.
// If the block doesn't exist an empty BlockLog will be passed to the function.
// TODO(jeremy): We should be able to replace this function with usage of dbutil.ReadModifyWrite
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
