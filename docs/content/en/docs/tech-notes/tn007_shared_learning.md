---
title: TN007 Shared Learning
description: A solution for sharing learning between multiple users
weight: 7
author: "[Jeremy Lewi](https://lewi.us/about)"
date: 2024-06-19
status: published
---

## Objective

* Design a solution for sharing learning between multiple users

## TL;DR

The examples used for in context learning are currently stored locally. This makes it awkward to share a trained
AI between team members. Users would have to manually swap the examples in order to benefit from each others learnings.
There are a couple key design decisions

* Do we centralize traces and block log events or only the learned examples?
* Do we support loading examples from a single location or multiple locations?

To simplify management I think we should do the following

* Only move the learned examples to a central location
* Trace/block logs should still be stored locally for each user
* Add support for loading/saving shared examples to multiple locations

Traces and block log events are currently stored in Pebble. Pebble doesn't have a good story for using shared
storage [cockroachdb/pebble#3177](https://github.com/cockroachdb/pebble/issues/3177#issuecomment-2137614459). 
We also don't have an immediate need to move the traces and block logs to a central location.

Treating shared storage location as a backup location means Foyle can still operate fine if the shared storage location
is inaccessible.

## Proposal

Users should be able to specify 

1. Multiple additional locations to load examples from
1. A backup location to save learned examples to

### LearnerConfig Changes

We can update LearnerConfig to support these changes

```
type LearnerConfig struct {
    // SharedExamples is a list of locations to load examples from
    SharedExamples []string `json:"sharedExamples,omitempty"`

    // BackupExamples is the location to save learned examples to
    BackupExamples string `json:"backupExamples,omitempty"`
}
```

### Loading SharedExamples

To support different storage systems (e.g. S3, GCS, local file system) we can define an interface for working
with shared examples. We currently have the [FileHelper](https://github.com/jlewi/hydros/blob/751cd2b5f0c7671f4e178c75292c55a9d827ecee/pkg/files/interface.go#L11)
interface

```go
type FileHelper interface {
    Exists(path string) (bool, error)
    NewReader(path string) (io.Reader, error)
    NewWriter(path string) (io.Writer, error)
}
```

Our current implementation of inMemoryDB requires a [Glob function](https://github.com/jlewi/foyle/blob/a811734050c23802e45b8d7a0031670c464c3971/app/pkg/learn/in_memory.go#L196)
to find all the examples that should be loaded. We should a new interface to include the Glob.

```go
type Globber interface {
    Glob(pattern string) ([]string, error)
}
```

For object storage we can implement Glob by listing all the objects matching a prefix and then applying the glob;
similar to this [code for matching a regex](https://github.com/jlewi/monogo/blob/c6693c86e89898f3a65c6f18b6b91b6e031c6dbd/gcp/gcs/util.go#L172)

### Triggering Loading of SharedExamples

For an initial implementation we can load shared examples when Foyle starts and perhaps periodically poll for
new examples. I don't think there's any need to implement push based notifications for new examples.


## Alternatives

### Centralize Traces and Block Logs

Since Pebble doesn't have a good story for using shared
storage [cockroachdb/pebble#3177](https://github.com/cockroachdb/pebble/issues/3177#issuecomment-2137614459) there's
no simple solution for moving the traces and block logs to a central location. 

The main thing we lose by not centralizing the traces and block is the ability to do bulk analysis of traces and
block events across all users. Since we don't have an immediate use case for that there's no reason to support it.

# References

[DuckDB S3 Support](https://duckdb.org/docs/extensions/httpfs/s3api.html#:~:text=DuckDB%20conforms%20to%20the%20S3,common%20among%20industry%20storage%20providers.)





