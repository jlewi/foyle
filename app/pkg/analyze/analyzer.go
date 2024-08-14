package analyze

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"

	"github.com/jlewi/foyle/app/pkg/docs"
	runnerv1 "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/runner/v1"

	"github.com/cockroachdb/pebble"
	"github.com/jlewi/foyle/app/api"
	"github.com/jlewi/foyle/app/pkg/dbutil"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
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

	// traceField is the field that contains the traceId in a log entry. We use this to identify processing related
	// to a particular trace. We don't use the field "traceId" because these log entries aren't actually part of the
	// trace.
	traceField = "targetTraceId"
)

// Analyzer is responsible for analyzing logs and building traces. It does this in a streaming fashion so that
// traces get built in "realtime".
//
// The Analyzer is multi-threaded. One potential pitfall is we have multiple writers trying to update the same
// key. This would result in a last-write win situation with the last write potentially overwriting the changes
// by the other writer. To avoid this, we use WorkQueue's to schedule updates to each key. The workqueue
// ensures that a single worker is processing a single key at a time. The items in the workqueue should be concrete
// types not pointers because the workqueues use the value as the key.
//
// If the Analyzer emits any info log messages in the course of processing Foyle's raw log entries; this will result
// in continual reprocessing of Foyle logs. This is because the Analyzer is watching the log file and fires whenever
// its updated and invoke handleLogFileEvents. So if handleLogFileEvents updates the log file then we get constant
// reprocessing. We use a rateLimitingQueue to ensure that we don't process the same file too quickly.
type Analyzer struct {
	tracesDB  *pebble.DB
	blocksDB  *dbutil.LockingDB[*logspb.BlockLog]
	rawLogsDB *dbutil.LockingDB[*logspb.LogEntries]
	// queue for log file processing
	queue workqueue.RateLimitingInterface
	// Queue for block log processing
	// TODO(jeremy): We should really use a durable queue backed by files
	blockQueue workqueue.DelayingInterface

	watcher *fsnotify.Watcher

	blockNotifier PostBlockEvent

	handleLogFileIsDone sync.WaitGroup
	handleBlocksIsDone  sync.WaitGroup
	logFileOffsets      map[string]int64
	mu                  sync.Mutex

	logOffsetsFile string
	// Only used during testing to allow the test to tell when the log file processing is done.
	signalFileDone  chan<- string
	signalBlockDone chan<- string
}

// NewAnalyzer creates a new Analyzer.
func NewAnalyzer(logOffsetsFile string, rawLogsDB *dbutil.LockingDB[*logspb.LogEntries], tracesDB *pebble.DB, blocksDB *dbutil.LockingDB[*logspb.BlockLog]) (*Analyzer, error) {
	logOffsets, err := initOffsets(logOffsetsFile)
	if err != nil {
		return nil, err
	}

	// Create a rate limiting queue for processing files. We rate limit to each file every 30 seconds. This is because
	// The logs are constantly being written to and we don't want to process the files too quickly.
	// We are potentially writing to multiple files at the same time e.g. the Analyzer logs and then a different
	// log file for each instance of RunMe. So we need to track different backoffs for each file which the rate limiter
	// does. Using exponential backoff would make sense when we update processLogFile to detect the end of a trace.
	// In that case, after we detect the start of a trace we would want to retry on a very short interval with backoff
	// to detect the end of the trace as quickly as possible. Right now we don't do that and in fact we never call
	// forget so we will basically max out the retry limit at the max delay.
	fileQueue := workqueue.NewRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(5*time.Second, 30*time.Second))

	return &Analyzer{
		logOffsetsFile: logOffsetsFile,
		rawLogsDB:      rawLogsDB,
		tracesDB:       tracesDB,
		blocksDB:       blocksDB,
		queue:          fileQueue,
		blockQueue:     workqueue.NewDelayingQueue(),
		logFileOffsets: logOffsets,
	}, nil
}

func initOffsets(logOffsetsFile string) (map[string]int64, error) {
	log := zapr.NewLogger(zap.L())
	raw, err := os.ReadFile(logOffsetsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]int64{}, nil
		}
		return nil, errors.Wrapf(err, "Failed to read watermarks file %s", logOffsetsFile)
	}
	watermarks := map[string]int64{}
	if err := json.Unmarshal(raw, &watermarks); err != nil {
		log.Error(err, "Failed to unmarshal watermarks file %s; watermarks will be reinitialized", logOffsetsFile)
	}
	return watermarks, nil
}

type fileItem struct {
	path string
}

type blockItem struct {
	id string
}

// PostBlockEvent interface for functions to post block events.
type PostBlockEvent func(id string) error

// Run runs the analyzer; continually processing logs.
// blockNotifier is an optional function that will be called when a block is updated.
// This should be non blocking.
func (a *Analyzer) Run(ctx context.Context, logDirs []string, blockNotifier PostBlockEvent) error {
	a.blockNotifier = blockNotifier
	// Find all the current files
	jsonFiles, err := findLogFilesInDirs(ctx, logDirs)
	if err != nil {
		return err
	}

	// Enqueue an item to process each file
	for _, f := range jsonFiles {
		a.queue.Add(fileItem{path: f})
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
			// N.B. We currently don't call forget on any files. So for each file we will max out the retry limit
			// at the max delay. It might make sense to call forget when we detect an open trace that we are waiting
			// on log entries to complete. In this case, we'd want to retry the file on a shorter interval with backoff
			// to detect the end of the trace with as little delay as possible.
			defer q.Done(item)
			fileItem, ok := item.(fileItem)
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
	log.V(logs.Debug).Info("Processing log file", "path", path)

	offset := a.getLogFileOffset(path)
	lines, offset, err := readLinesFromOffset(ctx, path, offset)
	if err != nil {
		return err
	}
	if len(lines) == 0 {
		return nil
	}

	traceIDs := make(map[string]bool)

	pkgPath := getFullPackagePath()
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

		// Drop all log entries that come from the Analyzer package itself. This should hadn't be neccessary
		// but its a precaution to guard against someone accidentally adding a log message with the the field "traceId"
		// to log a message about processing that trace. If we include such messages as part of the trace
		// we could trigger an infinite loop
		if strings.HasPrefix(entry.Function(), pkgPath) {
			log.Error(errors.New("Ignoring log entry from Analyzer package"), "Ignoring log entry from Analyzer package", "entry", entry)
		}

		if err := a.rawLogsDB.ReadModifyWrite(entry.TraceID(), func(entries *logspb.LogEntries) error {
			if entries.Lines == nil {
				entries.Lines = make([]string, 0, 1)
			}
			entries.Lines = append(entries.Lines, line)
			return nil
		}); err != nil {
			// If there is a problem writing to the DB we should probably surface it rather than just keep going.
			return err
		}

		traceIDs[entry.TraceID()] = true
	}

	// Combine the entries for each trace that we saw.
	// N.B. We could potentially make this more efficient by checking if the log message is the final message
	// in a trace. This would avoid potentially doing a combine for a trace on each log message.
	for tid := range traceIDs {
		if err := a.buildTrace(ctx, tid); err != nil {
			log.Error(err, "Error building trace", traceField, tid)
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

	log := zapr.NewLogger(zap.L())
	// Persist the watermarks
	raw, err := json.Marshal(a.logFileOffsets)
	if err != nil {
		log.Error(err, "Failed to marshal watermarks")
		return
	}
	// To do the write atomically we write to a temp file and then rename it.
	tempFile := fmt.Sprintf("%s.tmp", a.logOffsetsFile)
	if err := os.WriteFile(tempFile, raw, 0644); err != nil {
		log.Error(err, "Failed to write watermarks")
		return
	}

	if err := os.Rename(tempFile, a.logOffsetsFile); err != nil {
		log.Error(err, "Failed to rename watermarks file", "tempFile", tempFile, "logOffsetsFile", a.logOffsetsFile)
	}
	log.V(logs.Debug).Info("Wrote watermarks", "logOffsetsFile", a.logOffsetsFile)
}

// registerDirWatchers sets up notifications for changes in the log directories.
// Any time a file is modified it will enqueue the file for processing.
func (a *Analyzer) registerDirWatchers(ctx context.Context, q workqueue.RateLimitingInterface, logDirs []string) error {
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
func handleFsnotifications(ctx context.Context, watcher *fsnotify.Watcher, q workqueue.RateLimitingInterface) {
	log := logs.FromContext(ctx)
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				q.AddRateLimited(fileItem{path: event.Name})
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Error(err, "Error from watcher")
		}
	}
}

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

	log.Info("Analyzer shutdown")
	return nil
}

// buildTrace creates the trace and initializes the blocks.
func (a *Analyzer) buildTrace(ctx context.Context, tid string) error {
	log := logs.FromContext(ctx)

	// Entries is a mapping from a traceId to a list of logEntries associated with that entry.
	logEntries := make([]*api.LogEntry, 0, 10)

	protoEntries, err := a.rawLogsDB.Get(tid)
	if err != nil {
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
	log.Info("Combining entries for trace", traceField, tid, "numEntries", len(logEntries))
	trace, err := combineEntriesForTrace(ctx, logEntries)
	if err != nil {
		log.Error(err, "Error combining entries for trace", traceField, tid)
		return err
	}

	if trace == nil {
		// trace will be nil if the entries associated with the trace correspond to a type of trace that we currently
		// don't log in the traces DB. For example, right now we don't produce a trace for the streaming request
		log.V(logs.Debug).Info("Entries for trace are currently skipped", traceField, tid)
		return nil
	}

	if err := dbutil.SetProto(a.tracesDB, tid, trace); err != nil {
		return err
	}

	// Update the blocks associated with this trace; we need to update the block with any trace ids that it uses.
	// We will then enqueue the block for processing.
	blockIds, err := func() ([]string, error) {
		bids := make([]string, 0, 10)
		switch t := trace.Data.(type) {
		case *logspb.Trace_Generate:
			for _, oBlock := range t.Generate.Response.GetBlocks() {
				bid := oBlock.GetId()
				if bid == "" {
					continue
				}
				bids = append(bids, bid)
				if err := a.blocksDB.ReadModifyWrite(bid, func(block *logspb.BlockLog) error {
					block.Id = bid
					block.GenTraceId = tid
					return nil
				}); err != nil {
					return bids, errors.Wrapf(err, "Failed to set generate trace on block %s", bid)
				}
			}
		case *logspb.Trace_Execute:
			bid := t.Execute.Request.GetBlock().GetId()
			if bid == "" {
				return bids, nil
			}
			bids = append(bids, bid)
			if err := a.blocksDB.ReadModifyWrite(bid, func(block *logspb.BlockLog) error {
				block.Id = bid
				if block.ExecTraceIds == nil {
					block.ExecTraceIds = make([]string, 0, 10)
				}
				block.ExecTraceIds = append(block.ExecTraceIds, tid)
				return nil
			}); err != nil {
				return bids, errors.Wrapf(err, "Failed to set execute trace on block %s", bid)
			}
		case *logspb.Trace_RunMe:
			bid := t.RunMe.Request.GetKnownId()
			if bid == "" {
				return bids, nil
			}
			bids = append(bids, bid)
			if err := a.blocksDB.ReadModifyWrite(bid, func(block *logspb.BlockLog) error {
				block.Id = bid
				if block.ExecTraceIds == nil {
					block.ExecTraceIds = make([]string, 0, 10)
				}
				block.ExecTraceIds = append(block.ExecTraceIds, tid)
				return nil
			}); err != nil {
				return bids, errors.Wrapf(err, "Failed to set RunMe trace on block %s", bid)
			}
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
	for _, delta := range blockIds {
		a.blockQueue.Add(blockItem{
			id: delta,
		})
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

		return jsonFiles, errors.WithStack(fmt.Errorf("Analyze invoked for non-existent path: %v", logsDir))
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
			// N.B. We need to enqueue concrete types because the workqueue uses the value as the key.
			// If we use pointers then we would be using the address as the key and we will end up treating the same
			// values as different keys which would result in multiple workers processing the same item.
			blockItem, ok := item.(blockItem)
			if !ok {
				log.Error(errors.New("Failed to cast item to blockItem"), "Failed to cast item to blockItem")
				return
			}

			err := a.blocksDB.ReadModifyWrite(blockItem.id, func(block *logspb.BlockLog) error {
				return buildBlockLog(ctx, block, a.tracesDB)
			})
			if err != nil {
				log.Error(err, "Error processing block", "block", blockItem.id)
			}
			if a.blockNotifier != nil {
				if err := a.blockNotifier(blockItem.id); err != nil {
					log.Error(err, "Error notifying block event", "block", blockItem.id)
				}
			}
			if a.signalBlockDone != nil {
				a.signalBlockDone <- blockItem.id
			}
		}()
	}
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

	// Dedupe the execution traces just in case
	uEids := make(map[string]bool)
	for _, eid := range block.GetExecTraceIds() {
		uEids[eid] = true
	}
	block.ExecTraceIds = make([]string, 0, len(uEids))
	for eid := range uEids {
		block.ExecTraceIds = append(block.ExecTraceIds, eid)
	}

	eidToTime := make(map[string]time.Time)

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

			eidToTime[tid] = trace.StartTime.AsTime()

			if lastTrace == nil {
				lastTrace = trace
				return
			}

			if lastTrace.StartTime.AsTime().Before(trace.StartTime.AsTime()) {
				lastTrace = trace
			}
		}()
	}

	// Sort execTrace ids based on their time. This is so the ordering is stable for the unittest.
	// It should also be convenient for manual analysis since we usually care about the last exec trace.
	sort.Slice(block.ExecTraceIds, func(i, j int) bool {
		left := block.ExecTraceIds[i]
		right := block.ExecTraceIds[j]
		leftTime := eidToTime[left]
		rightTime := eidToTime[right]
		return leftTime.Before(rightTime)
	})

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

		if strings.HasSuffix(function, "agent.(*Agent).StreamGenerate") {
			// For now we do nothing with StreamGenerate traces.
			return nil, nil
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

		if gTrace.Request == nil && strings.HasSuffix(e.Function(), "agent.(*Agent).Generate") {
			raw := e.Request()
			if raw != nil {
				request := &v1alpha1.GenerateRequest{}
				if err := protojson.Unmarshal([]byte(raw), request); err != nil {
					return nil, errors.Wrapf(err, "Failed to unmarshal GenerateRequest for trace %s", trace)
				}

				gTrace.Request = request
				trace.StartTime = timestamppb.New(e.Time())
			}
		}
		if gTrace.Response == nil && strings.HasSuffix(e.Function(), "agent.(*Agent).Generate") {
			raw := e.Response()
			if raw != nil {
				v := &v1alpha1.GenerateResponse{}
				if err := protojson.Unmarshal([]byte(raw), v); err != nil {
					return nil, errors.Wrapf(err, "Failed to unmarshal GenerateResponse for trace %s", trace.Id)
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
