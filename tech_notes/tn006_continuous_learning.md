# Continuous Learning

* **Author**: Jeremy Lewi
* **Last Updated**: 2024-06-07
* **Status**: Published

## Objective

* Continuously learn from the traces and block logs

## TL;DR

[TN005 Continuous Log Processing](tn005_continuous_log_processing.md) 
described how we can continuously process logs. This technote describes how we can continuously learn from the logs.

## Background

### How Learning Works

[TN002 Learning](tn002_learning.md) provided the initial design for learning from human feedback.

This was implemented in [Learner.Reconcile](https://github.com/jlewi/foyle/blob/cfc76ecdb252a73b4f512e1022a7ef61cc4321cb/app/pkg/learn/learner.go#L46)
as two reconcilers

* `reconcile.reconcileExamples` - Iterates over the blocks database and for each block produces a training example if one is
  required

* `reconcile.reconcileEmbeddings` - Iterates over the examples and produces embeddings for each example

Each block currently produces at most 1 training example and the block id and the example ID are the same.

[InMemoryDB](https://github.com/jlewi/foyle/blob/cfc76ecdb252a73b4f512e1022a7ef61cc4321cb/app/pkg/learn/in_memory.go#L23)
implements RAG using an in memory database. 

## Proposal

We can change the signature of `Learner.Reconcile` to be 

```
func (l *Learner) Reconcile(ctx context.Context, id string) error
```

where id is a unique identifier for the block. We can also update Learner to use a workqueue to process blocks
asynchronously.

```
func (l *Learner) Start(ctx context.Context, events <-chan string, updates chan<-string) error
```

The learner will listen for events on the event channel and when it receives an event it will enqueue the example
for reconciliation. The updates channel will be used to notify `InMemoryDB` that it needs to load/reload an example.

For actual implementation we can simplify `Learning.reconcileExamples` and `Learning.reconcileEmbeddings`
since we can eliminate the need to iterate over the entire database.

### Connecting to the Analyzer

The analyzer needs to emit events when there is a block to be added.  We can change the signature of the run function
to be 

```
func (a *Analyzer) Run(ctx context.Context, notifier func(string)error ) error
```

The notifier function will be called whenever there is a block to be added. Using a function means the Analyzer
doesn't have to worry about the underlying implementation (e.g. channel, workqueue, etc.).

### Backfilling

We can support backfilling by adding a `Backfill` method to `Learner` which will iterate over the blocks database.
This is useful if we need to reprocess some blocks because of a bug or because of an error.

```
func (l *Learner) Backfill(ctx context.Context, startTime time.time) error
```

We can use `startTime` to filter down the blocks that get reprocessed. Only those blocks that were modified
after start time will get reprocessed.

## InMemoryDB

We'd like to update `InMemoryDB` whenever an example is added or updated. We can use the `updates` channel
to notify `InMemoryDB` that it needs to load/reload an example.

Internally `InMemoryDB` uses a matrix and an array to store the data

* embeddings stores the embeddings for the examples in a num_examples x num_features.
* embeddings[i, :] is the embedding for examples[i]

We'll need to add a map from example ID to the row in the embeddings matrix `map[string]int`. We'll also
need to protect the data with a mutex to make it thread safe.
