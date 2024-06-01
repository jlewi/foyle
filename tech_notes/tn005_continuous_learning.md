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

### Continuous Learning

When it comes to continuous learning we have to potential options

1. We could compute any examples for learning as part of processing a blocklog
2. We could have a separate queue for learning and add events as a result of processing a blocklog

I think it makes sense to keep learning as a separate step. The learning process will likely evolve over time and
its advantageous if we can redo learning without having to reprocess the logs.









