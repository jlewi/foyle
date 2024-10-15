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

	"github.com/jlewi/foyle/app/pkg/logs/matchers"

	"github.com/jlewi/foyle/app/pkg/runme/converters"
	parserv1 "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/parser/v1"
	"google.golang.org/protobuf/proto"

	"github.com/fsnotify/fsnotify"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"

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

	learnNotifier PostBlockEvent

	handleLogFileIsDone sync.WaitGroup
	handleBlocksIsDone  sync.WaitGroup
	logFileOffsets      *logspb.LogsWaterMark
	mu                  sync.Mutex

	logOffsetsFile string
	// Only used during testing to allow the test to tell when the log file processing is done.
	signalFileDone  chan<- string
	signalBlockDone chan<- string

	sessBuilder *sessionBuilder
}

// NewAnalyzer creates a new Analyzer.
func NewAnalyzer(logOffsetsFile string, rawLogsDB *dbutil.LockingDB[*logspb.LogEntries], tracesDB *pebble.DB, blocksDB *dbutil.LockingDB[*logspb.BlockLog], sessions *SessionsManager) (*Analyzer, error) {
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

	sessBuilder, err := NewSessionBuilder(sessions)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create session builder")
	}

	return &Analyzer{
		logOffsetsFile: logOffsetsFile,
		rawLogsDB:      rawLogsDB,
		tracesDB:       tracesDB,
		blocksDB:       blocksDB,
		queue:          fileQueue,
		blockQueue:     workqueue.NewDelayingQueue(),
		logFileOffsets: logOffsets,
		sessBuilder:    sessBuilder,
	}, nil
}

func initOffsets(logOffsetsFile string) (*logspb.LogsWaterMark, error) {
	log := zapr.NewLogger(zap.L())
	raw, err := os.ReadFile(logOffsetsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &logspb.LogsWaterMark{}, nil
		}
		return nil, errors.Wrapf(err, "Failed to read watermarks file %s", logOffsetsFile)
	}
	watermark := &logspb.LogsWaterMark{}

	if err := protojson.Unmarshal(raw, watermark); err != nil {
		log.Error(err, "Failed to unmarshal watermarks file %s; watermarks will be reinitialized", logOffsetsFile)
	}
	return watermark, nil
}

type fileItem struct {
	path string

	// Whether the file is still active. If the file is no longer active we should stop processing it.
	active bool
}

type blockItem struct {
	id string
}

// PostBlockEvent interface for functions to post block events.
type PostBlockEvent func(id string) error

// Run runs the analyzer; continually processing logs.
// learnNotifier is an optional function that will be called when a block is updated.
// This should be non blocking.
func (a *Analyzer) Run(ctx context.Context, logDirs []string, learnNotifier PostBlockEvent) error {
	a.learnNotifier = learnNotifier
	// Find all the current files
	jsonFiles, err := findLogFilesInDirs(ctx, logDirs)
	if err != nil {
		return err
	}

	// Enqueue an item to process each file
	for i, f := range jsonFiles {
		// Only the last file should be active.
		active := i == len(jsonFiles)-1
		a.queue.Add(fileItem{path: f, active: active})
	}

	a.handleLogFileIsDone.Add(1)
	a.handleBlocksIsDone.Add(1)

	// Important we should only process LogFileEvents in a single go func because the semantics of the watermark
	// are that all log entries up to that mark have been processed. If we process log entries in parallel its not
	// clear how we handle updating the waterMark.
	go a.handleLogFileEvents(ctx)
	go a.handleBlockEvents(ctx)

	return nil
}

// TODO(jeremy): How do we make the Analyzer thread safe? I believe the DB classes are thread safe
//
// TODO(jeremy): Should we instrument this with OTEL to get metrics on how long it takes to process a log file?
// What we'd like is counters for how often a log file is processed. But maybe we should use logs for that?
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
			// If the file is still active re-enqueue it.
			if fileItem.active {
				q.AddRateLimited(fileItem)
			} else {
				log.Info("Finished processing log file", "path", fileItem.path)
			}

		}()
	}
}

func (a *Analyzer) processLogFile(ctx context.Context, path string) error {
	log := logs.FromContext(ctx)
	log.V(logs.Debug).Info("Processing log file", "path", path)

	offset := a.getLogFileOffset(path)
	if offset <= -1 {
		// Offset of -1 means we are done processing the file because it is before the watermark
		log.V(logs.Debug).Info("Logfile already processed", "path", path)
		return nil
	}

	maxLines := 200

	// TODO(jeremy): We use pkgPath to filter out log entries from the Analyzer package.
	// We could use the pattern illustrated by the fnames package of using a constant to define the package path
	// and then a unittest which uses reflection to verify the constant is correct.
	pkgPath := getFullPackagePath()

	for {
		// Keep reading lines from the file until we reach the end.
		// We process the log entries in chunks of maxLines. After every maxLines read we will perform checkpointing.
		// This is to ensure that when backfilling we make progress
		var err error
		var lines []string
		// n.b. if we do lines,offset, err := we will end up shadowing offset and on each call to readLinesFromOffset
		// the value of offset won't be the new value
		lines, offset, err = readLinesFromOffset(ctx, path, offset, maxLines)

		if err != nil {
			return err
		}
		if len(lines) == 0 {
			return nil
		}

		traceIDs := make(map[string]bool)

		// We read the lines line by line. We keep track of all the traceIDs mentioned in those lines. We
		// Then do a combineAndCheckpoint for all the traceIDs mentioned. Lines are also persisted in a KV store
		// keyed by traceID. So if on the next iteration we get a new line for a given traceId and need to reprocess
		// the trace we can do that because we can fetch all the line entries for that trace.
		for _, line := range lines {
			entry := &api.LogEntry{}
			if err := json.Unmarshal([]byte(line), entry); err != nil {
				log.Error(err, "Error decoding log entry", "path", path, "line", line)
				continue
			}

			// Add the entry to a session if it should be.
			a.sessBuilder.processLogEntry(entry)

			if matchers.IsLogEvent(entry.Function()) {
				a.processLogEvent(ctx, entry)
				continue
			}

			// Ignore log entries without traces
			if entry.TraceID() == "" {
				continue
			}

			// Drop all log entries that come from the Analyzer package itself. This shouldn't be neccessary
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
				log.Error(err, "Failed to write log entry to DB", "entry", entry)
				continue
			}

			traceIDs[entry.TraceID()] = true
		}

		// Now run a combineAndCheckpoint
		a.combineAndCheckpoint(ctx, path, offset, traceIDs)

		// If we are shutting down we don't want to keep processing the file.
		// By aborting shutdown here as opposed to here we are blocking shutdown for as least as long it takes
		// to process maxLines. If maxLines is large it could be a while.
		if a.queue.ShuttingDown() {
			log.Info("Halting processing of log file because Analyzer is shutting down", "path", path)
			return nil
		}
	}
}

// combineAndCheckpoint runs a combine operation for all the traceIDs listed in the map.
// Progress is then checkpointed.
func (a *Analyzer) combineAndCheckpoint(ctx context.Context, path string, offset int64, traceIDs map[string]bool) {
	log := logs.FromContext(ctx)
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
}

func (a *Analyzer) GetWatermark() *logspb.LogsWaterMark {
	a.mu.Lock()
	defer a.mu.Unlock()
	w := proto.Clone(a.logFileOffsets).(*logspb.LogsWaterMark)
	return w
}

// getLogOffSet returns the offset for the log file to start reading from.
// A value < 0 means the watermark is already past the end of the file and no more processing is needed.
func (a *Analyzer) getLogFileOffset(path string) int64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	// N.B. This code takes into account the full file path when deciding the ordering of the logfiles.
	if path < a.logFileOffsets.File {
		return -1
	}
	if path > a.logFileOffsets.File {
		return 0
	}
	return a.logFileOffsets.Offset
}

func (a *Analyzer) setLogFileOffset(path string, offset int64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	oldWatermark := a.logFileOffsets
	a.logFileOffsets = &logspb.LogsWaterMark{
		File:   path,
		Offset: offset,
	}

	log := logs.NewLogger()
	if path < oldWatermark.File {
		log.Error(errors.New("Watermark is moving backwards"), "Watermark is moving backwards", zap.Object("oldWatermark", oldWatermark), zap.Object("newWatermark", a.logFileOffsets))
	}

	if oldWatermark.File != a.logFileOffsets.File {
		log.Info("Logs watermark moving to new file", zap.Object("oldWatermark", oldWatermark), zap.Object("newWatermark", a.logFileOffsets))
	}

	// Persist the watermarks
	raw, err := protojson.Marshal(a.logFileOffsets)
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

// processLogEvent processes a log event.
func (a *Analyzer) processLogEvent(ctx context.Context, entry *api.LogEntry) {
	log := logs.FromContext(ctx)

	event := &v1alpha1.LogEvent{}

	if !entry.GetProto("event", event) {
		log.Error(errors.New("Failed to decode event"), "Failed to decode LogEvent", "entry", entry)
		return
	}
	log = log.WithValues("eventId", event.GetEventId())
	switch event.Type {
	case v1alpha1.LogEventType_EXECUTE:
		bid := event.SelectedId
		if bid == "" {
			log.Error(errors.New("No selectedId"), "Execute event is missing selected id", "event", event)
			return
		}

		var cell *parserv1.Cell
		for _, c := range event.GetCells() {
			if converters.GetCellID(c) == bid {
				cell = c
				break
			}
		}

		if cell == nil {
			log.Error(errors.New("Failed to find cell"), "Execution log event is missing the actual cell", "bid", bid, "event", event)
			return
		}
		executedBlock, err := converters.CellToBlock(cell)
		if err != nil {
			jb, err := protojson.Marshal(cell)
			if err != nil {
				log.Error(err, "Failed to convert executed cell to block", "cellId", bid, "cell", string(jb))
			} else {
				log.Error(err, "Failed to convert executed cell to block", "cellId", bid)
			}
		}

		if err := a.blocksDB.ReadModifyWrite(bid, func(block *logspb.BlockLog) error {
			block.Id = bid
			block.ExecutedBlock = executedBlock
			return nil
		}); err != nil {
			log.Error(err, "Failed to update block with execution", "blockId", bid)
		}
		// We need to enqueue the block for processing since it was executed.
		// The learner will decide whether the blockLog has all the information it needs otherwise it will
		// disregard the block item and wait for further events.
		if a.learnNotifier != nil {
			if err := a.learnNotifier(bid); err != nil {
				log.Error(err, "Error notifying block event", "blockId", bid)
			}
		}
	case v1alpha1.LogEventType_ACCEPTED:
		fallthrough
	case v1alpha1.LogEventType_REJECTED:
		status := logspb.SuggestionStatus_SuggestionStatusUnknown
		switch event.Type {
		case v1alpha1.LogEventType_ACCEPTED:
			status = logspb.SuggestionStatus_ACCEPTED
		case v1alpha1.LogEventType_REJECTED:
			status = logspb.SuggestionStatus_REJECTED
		}

		for _, c := range event.GetCells() {
			bid := converters.GetCellID(c)
			if bid == "" {
				log.Error(errors.New("No cell id"), "Cell is missing id", zap.Object("event", event))
				continue
			}

			if err := a.blocksDB.ReadModifyWrite(bid, func(block *logspb.BlockLog) error {
				block.Id = bid
				block.SuggestionStatus = status
				return nil
			}); err != nil {
				log.Error(err, "Failed to update block with execution", "blockId", bid)
			}
		}
	default:
		// Do Nothing with the event
	}
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
				log.Error(err, "Error processing block", "blockId", blockItem.id)
			}
			// We need to enqueue the block for processing since it was executed.
			// The learner will decide whether the blockLog has all the information it needs otherwise it will
			// disregard the block item and wait for further events.
			if a.learnNotifier != nil {
				if err := a.learnNotifier(blockItem.id); err != nil {
					log.Error(err, "Error notifying block event", "blockId", blockItem.id)
				}
			}
			if a.signalBlockDone != nil {
				a.signalBlockDone <- blockItem.id
			}
		}()
	}
}

// buildBlockLog updates blocklogs given a generate trace.
// Since a single generate trace can generate multiple blocks, its a one to many operation.
func buildBlockLog(ctx context.Context, block *logspb.BlockLog, tracesDB *pebble.DB) error {
	log := logs.FromContext(ctx)
	log = log.WithValues("blockId", block.Id)
	log.Info("Building block log")

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

		if strings.HasSuffix(function, "agent.(*Agent).StreamGenerate") {
			// For now we do nothing with StreamGenerate traces.
			return nil, nil
		}
	}

	return nil, errors.New("Failed to identify trace type")
}

func combineGenerateTrace(ctx context.Context, entries []*api.LogEntry) (*logspb.Trace, error) {
	log := logs.FromContext(ctx)
	gTrace := &logspb.GenerateTrace{}
	trace := &logspb.Trace{
		Data: &logspb.Trace_Generate{
			Generate: gTrace,
		},
		Spans:      make([]*logspb.Span, 0, 10),
		Assertions: make([]*v1alpha1.Assertion, 0),
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

		if e.Message() == logs.Level1Assertion {
			assertion := &v1alpha1.Assertion{}
			if !e.GetProto("assertion", assertion) {
				log.Error(errors.New("Failed to decode assertion"), "Failed to decode assertion", "entry", e)
				continue
			}
			trace.Assertions = append(trace.Assertions, assertion)
			continue
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
