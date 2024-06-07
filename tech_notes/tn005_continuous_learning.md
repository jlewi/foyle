# Continuous Log Processing and Learning

* **Author**: Jeremy Lewi
* **Last Updated**: 2024-05-31
* **Status**: Being Drafted

## Objective

* Continuously compute traces and block logs
* Continuously learn from the traces and block logs

## TL;DR

Right now the user has to explicitly run

  1. `foyle log analyze` - to compute the aggregated traces and block logs
  1. `foyle learn` - to learn from the aggregated traces and block logs

This has a number of drawbacks that we'd like to fix [jlewi/foyle#84](https://github.com/jlewi/foyle/issues/84)

* Traces aren't immediately available for debugging and troubleshooting
* User doesn't immediately and automatically benefit from fixed examples
* Server needs to be stopped to analyze logs and learn because pebble can only be used 
  from a single process [jlewi/foyle#126](https://github.com/jlewi/foyle/issues/126)

We'd like to fix this so that logs are continuously processed and learned from as a background
process in the Foyle application. 

## Background

### Foyle Logs

Each time the Foyle server starts it creates a new timestamped log file. 
Currently, we only expect one Foyle server to be running at a time. 
Thus, at any given time only the latest log file might still be open for additional writes.

### RunMe Logs

RunMe launches multiple instances of the RunMe gRPC server. I think its one per vscode workspace.
Each of these instances writes logs to a separate timestamped file.
Currently, we don't have a good mechanism to detect when a log file has been closed and will recieve no more
writes.

### Block Logs

Computing the block logs requires doing two joins

* We first need to join all the log entries by their trace ID
* We then need to key each trace by the block IDs associated with that
* We need to join all the traces related to a block id 

### File Backed Queue Based Implementations

[rstudio/filequeue](https://github.com/rstudio/filequeue/blob/main/filequeue.go) implements a FIFO using a file per item 
and relies on a timestamp in the filename to maintain ordering. It renames files to pop them from the queue.

[nsqio/go-diskqueue](https://github.com/nsqio/go-diskqueue/blob/master/diskqueue.go) uses numbered files to implement 
a FIFO queue. A metadata file is used to track the position for reading and writing. The library automatically
rolls files when they reach a max size. Writes are asynchronous. Filesystem syncs happen periodically or whenever
a certain number of items have been read or written. If we set the sync count to be one it will sync after every write.


## Proposal 

### Accumulating Log Entries

We can accumulate log entries keyed by their trace ID in our KV store; for the existing local implementation this will be 
PebbleDB. In response to a new log entry we can fire an event to trigger a reduce operation on the log entries for that 
trace ID. This can in turn trigger a reduce operation on the block logs.

We need to track a watermark for each file so we know what entries have been processed. We can use the following
data structure

```go
// Map from a file to its watermark
type Watermarks map[string]Watermark

type Watermark struct {  
  // The offset in the file to continue reading from.
  Offset int64
}
```

If a file is longer than the offset then there's additional data to be processed. We can use 
[fsnotify](https://github.com/fsnotify/fsnotify) to get notified when a file has changed.

We could thus handle this as follows
* Create a FIFO where every event represents a file to be processed
* On startup register a fsnotifier for the directories containing logs
* Fire a sync event for each file when it is modified
* Enqueue an event for each file in the directories 
* When processing each file read its watermark to determine where to read from
* Periodically persist the watermarks to disk and persist on shutdownTo make this work need to implement a queue with a watermark. 

An in memory FIFO is fine because the durability comes from persisting the watermarks. If the watermarks
are lost we would reprocess some log entries but that is fine. We can reuse Kubernetes 
[WorkQueue](https://pkg.go.dev/k8s.io/client-go/util/workqueue)  to get "stinginess"; i.e 
avoid processing a given file multiple times concurrently and ensuring we only process a given item once if
it is enqued multiple times before it can be processed.

The end of a given trace is marked by specific known log entries e.g. 
["returning response"](https://github.com/jlewi/foyle/blob/e2da53a6e1c04f6bd87414c3aa8c2e37d3cac6c1/app/pkg/agent/agent.go#L114)
so we can trigger accumulating a trace when those entries are seen.

The advantage of this approach is we can avoid needing to create another durable, queue to trigger
trace processing because we can just rely on the watermark for the underlying log entry. In effect,
those watermarks will track the completion of updating the traces associated with any log entries up to
the watermark.

We could also use this trace ending messages to trigger garbage collection of the raw log entries in our KV store.

#### Implementation Details

* Responsibility for opening up the pebble databases should move into our application class 
  * This will allow db references to be passed to any classes that need them

* We should define a new DB to store the raw log entries
  * We should define a new proto to represent the values
* Analyzer should be changed to continuously process the log files
  * Create a FIFO for log file events
  * Persist watermarks to disk
  * Register a fsnotifier for the directories containing logs

### BlockLogs

When we perform a reduce operation on the log entries for a trace we can emit an event for any block logs that need to be
updated. We can enqueue these in a durable queue using 
[nsqio/go-diskqueue](https://github.com/nsqio/go-diskqueue/blob/master/diskqueue.go).

We need to accumulate (block -> traceIds[]string). We need to avoid multiple writers trying to update
the same block concurrently because that would lead to a last one wins situation. 

One option would be to have a single writer for the blocks database. Then we could just use a queue
for different block updates. Downside here is we would be limited to a single thread processing all block updates.

An improved version of the above would be to have multiple writers but ensure a given block can only be processed
by one worker at a time. We can use something like [workqueue](https://pkg.go.dev/k8s.io/client-go/util/workqueue)
for this. Workqueue alone won't be sufficient because it doesn't let you attach a payload to the enqueued item. The 
enqueued item is used as the key. Therefore we'd need a separate mechanism to keep track of all the updates that
need to be applied to the block.

An obvious place to accumulate updates to each block would be in the blocks database itself. Of course that brings
us back to the problem of ensuring only a single writer to a given block at a time. We'd like to make it easy to
for code to supply a function that will apply an update to a record in the database.

```go
func ReadModifyWrite[T proto.Message](db *pebble.DB, key string, msg T, modify func(T) error) error {
	...
}
```

To make this thread safe we need to ensure that we never try to update the same block concurrently. We can do
that by implementing row level locks. [fishy/rowlock](https://github.com/fishy/rowlock/blob/master/rowlock.go) is
an example. It is basically, a locked map of row keys to locks. Unfortunately, it doesn't implement any type of
forgetting so we'd keep accumulating keys. I think we can use the builtin [sync.Map](https://pkg.go.dev/sync#Map)
to implement RowLocking with optimistic concurrency. The semantics would be like the following

* Add a [ResourceVersion](https://github.com/kubernetes/apimachinery/blob/703232ea6da48aed7ac22260dabc6eac01aab896/pkg/apis/meta/v1/types.go#L172)
  to the proto that can be used for optimistic locking
* Read the row from the database
* Set ResourceVersion if not already set
* Call LoadOrStore to load or store the resource version in the sync.Map
  * If a different value is already stored then restart the operation
* Apply the update
* Generate a new resource version
* Call CompareAndSwap to update the resource version in the sync.Map
  * If the value has changed then restart the operation
* Write the updated row to the database
* Call CompareAndDelete to remove the resource version from the sync.Map

### Continuous Learning

When it comes to continuous learning we have to potential options

1. We could compute any examples for learning as part of processing a blocklog
2. We could have a separate queue for learning and add events as a result of processing a blocklog

I think it makes sense to keep learning as a separate step. The learning process will likely evolve over time and
its advantageous if we can redo learning without having to reprocess the logs.









