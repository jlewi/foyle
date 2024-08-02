---
title: TN001 Logging
description: Design logging to support capturing human feedback.
weight: 1
author: "[Jeremy Lewi](https://lewi.us/about)"
date: 2024-04-10
---

* **Status**: Being Drafted

## Objective

Design logging to support capturing human feedback.

## TL;DR

One of the key goals of Foyle is to collect human feedback to improve the quality of the AI.
The key interaction that we want to capture is as follows

* User asks the AI to generate one or more commands
* User edits those commands
* User executes those commands

In particular, we want to be able to link the commands that the AI generated to the commands that the user ultimately
executes. We can do this as follows

* When the AI generates any blocks it attaches a unique block id to each block
* When the frontend makes a request to the backend, the request includes the block ids of the blocks
* The blockid can then be used as a join key to link what the AI produced with what a human ultimately executed

## Implementation

### Protos

We need to add a new field to the `Block` proto to store the block id.

```proto
message Block {
   ...
    string block_id = 7;
}
```

As of right now, we don't attach a block id to the `BlockOutput` proto because

1. We don't have an immediate need for it
2. Since `BlockOutput` is a child of a `Block` it is already linked to an id

### Backend

On the backend we can rely on structured logging to log the various steps (e.g. RAG) that go into producing a block.
We can add log statements to link a block id to a traceId so that we can see how that block was generated.

#### Logging Backend

When it comes to a logging backend the simplest most universal solution is to write the structured logs as JSONL to a 
file. The downside of this approach is that this schema wouldn't be efficient for querying. We could solve that by
choosing a logging backend that supports indexing. For example, SQLite or Google Cloud Logging. I think it will
be simpler to start by appending to a file. Even if we end up adding SQLite it might be advantageous to have a separate
ETL pipeline that reads the JSONL and writes it to SQLite. This way each entry in the SQLite database could potentially
be a trace rather than a log entry.

To support this we need to modify `App.SetupLogging` to log to the appropriate file.

I don't think we need to worry about log rotation. I think we can just generate a new timestamped file each time Foyle
starts. In cases where Foyle might be truly long running (e.g. deployed on K8s) we should probably just be logging
to stodut/stderr and relying on existing mechanisms (e.g. Cloud Logging to ingest those logs).

### Logging Delete Events

We might also want to log delete block events. If the user deletes an AI generated block that's a pretty strong signal
that the AI made a mistake; for example providing an overly verbose command. We could log these events by adding
a VSCode handler for the delete event. We could then add an RPC to the backend to log these events. I'll probably
leave this for a future version.

## Finding Mistakes

One of the key goals of logging is to find mistakes. In the case of Foyle, a mistake is where the AI generated block
got edited by the human before being executed. We can create a simple batch job (possibly using Dataflow) to find
these examples. The job could look like the following

* Filter the logs to find entries and emit tuples corresponding to AI generated block (block_id, trace_id, contentents)
  and executed blocks (block_id, trace_id, contents)
* Join the two streams on block_id
* Compare the contents of the two blocks if they aren't the same then it was a mistake 

## References

[Distributed Tracing](https://docs.honeycomb.io/get-started/basics/observability/concepts/distributed-tracing/)


