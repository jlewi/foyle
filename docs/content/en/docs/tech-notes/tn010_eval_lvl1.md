---
title: TN010 Level 1 Evaluation
description: Level 1 Evaluation
weight: 10
author: "[Jeremy Lewi](https://lewi.us/about)"
date: 2024-08-02
status: Being Drafted
---

# Objective:

Design level 1 evaluation to optimize responses for AutoComplete

# TL;DR

As we roll out [AutoComplete](../tn008_auto_insert_cells) we are observing
that AI quality is an issue [jlewi/foyle#170](https://github.com/jlewi/foyle/issues/170).
Examples of quality issues we are seeing are

* AI suggests a code cell rather than a markdown cell when a user is editing a markdown cell
* AI splits commands across multiple code cells rather than using a single code cell

We can catch these issues using [Level 1 Evals](https://hamel.dev/blog/posts/evals/#level-1-unit-tests)
which are basically assertions applied to the AI responses.

To implement Level 1 Evals we can do the following

1. We can generate an evaluation from logs of actual user interactions
1. We can create scripts to run the assertions on the AI responses
1. We can use RunMe to create playbooks to run evaluations and visualize the results as well as data

## Background: 

### Existing Eval Infrastructure

[TN003](../tn003_learning_eval/) described how we could do evaluation given a golden dataset of examples.
The [implementation](https://github.com/jlewi/foyle/tree/main/app/pkg/eval) was motivated by a number of factors.

We want a resilient design because batch evaluation is a long running, flaky process because it depends on an external 
service (i.e. the LLM provider). Since LLM invocations cost money we want to avoid needlessly recomputing 
generations if nothing had changed. Therefore, we want to be able to checkpoint progress and resume from where we left off. 

We want to run the same codepaths as in production. Importantly, we want to reuse logging and visualization tooling
so we can inspect evaluation results and understand why the AI succeeded. 

To run experiments, we need to modify the code and deploy a new instance of Foyle just for evaluation.
We don't want the evaluation logs to be mixed with the production logs; in particular we don't want to learn from
evaluation data.

Evaluation was designed around a controller pattern. The controller figures out which computations need to be run
for each data point and then runs them. Thus, once a data point is processed, rerunning evaluation becomes a null-op.
For example, if an LLM generation has already been computed, we don't need to recompute it. Each data point
was an instance of the [EvalResult Proto](https://github.com/jlewi/foyle/blob/f71718a50ce131a464d884f97bd0de18c24bafc5/protos/foyle/v1alpha1/eval.proto#L17). A Pebble Database was used to store the results. 

Google Sheets was used for reporting of results.

Experiments were defined using the [Experiment Resource](https://github.com/jlewi/foyle/blob/main/app/api/experiment.go) via [YAML files](https://github.com/jlewi/foyle/blob/f71718a50ce131a464d884f97bd0de18c24bafc5/experiments/rag.yaml). An experiment could
be run using the Foyle CLI.

## Defining Assertions

As noted in [Your AI Product Needs Evals](https://hamel.dev/blog/posts/evals/#level-1-unit-tests) we'd like
to design our assertions so they can be run online and offline. 

We can use the following interface to define assertions

```go
type Assertion interface {  
  Assert(ctx context.Context, doc *v1alpha1.Doc, examples []*v1alpha1.Example, answer []*v1alpha1.Block) (AssertResult, error)
  // Name returns the name of the assertion.
  Name() string
}

type AssertResult string
AssertPassed AssertResult = "passed"
AssertFailed AssertResult = "failed"
AssertSkipped AssertResult = "skipped"
```

The `Assert` method takes a document, examples, and the AI response and returns a triplet indicating whether the assertion passed
or was skipped. Context can be used to pass along the `traceId` of the actual request.

### Online Evaluation

For online execution, we can run the assertions asynchronously in a thread. We can log the assertions using existing logging patterns. This will allow us to fetch the assertion results as part of the trace. Reporting the results should not be the responsibility of each Assertion; we should handle that centrally. We will use OTEL to report the results as well; each 
assertion will be added as an attribute to the trace. This will make it easy to monitor performance over time.

## Batch Evaluation

For quickly iterating on the AI, we need to be able to do offline, batch evaluation. Following
the existing patterns for evaluation, we can define a new `AssertJob` resource to run the assertions.

```yaml
kind: AssertJob
apiVersion: foyle.io/v1alpha1
metadata:
  name: "learning"
spec:
  sources:
    - traceServer:
        address: "http://localhost:8080"
    - mdFiles:
        path: "foyle/evalData"

  # Pebble database used to store the results
  dbDir: /Users/jlewi/foyle_experiments/20250530-1612/learning
  agent: "http://localhost:108080"
  sheetID: "1iJbkdUSxEkEX24xMH2NYpxqYcM_7-0koSRuANccDAb8"
  sheetName: "WithRAG"  
```

The `source` field specifies the sources for the evaluation dataset. There are two different kinds of sources. A traceServer
specifies an instance of a Foyle Agent that makes its traces available via an API. This can be used to generate examples
based on actual user interactions. The `Traces` need to be read via API and not directly from the pebble database because the pebble database is not designed for concurrent access. 

The `mdFiles` field allows examples to be provided as markdown files. This will be used to create a handcrafted curated
dataset of examples.

The `Agent` field specifies the address of the instance of Foyle to be evaluated. This instance should be configured to store its data in a different location. 

The `SheetID` and `SheetName` fields specify the Google Sheet where the results will be stored.

To perform the evaluation, we can implement a controller modeled on our existing [Evaluator](https://github.com/jlewi/foyle/blob/main/app/pkg/eval/evaluator.go).

## Traces Service

We need to introduce a Trace service to allow the evaluation to access traces.

```proto

service TracesService {
  rpc ListTraces(ListTracesRequest) returns (ListTracesResponse) {
  }
}
```

We'll need to support filtering and pagination. The most obvious way to filter would be on time range.
A crude way to support time based filtering would be as follows
* Raw log entries are written in timestamp order
* Use the raw logs to read log entries based on time range
* Get the unique set of TraceIDs in that time range
* Look up each trace in the traces database

## Initial Assertions

Here are some initial assertions we can define

* If human is editing a markdown cell, suggestion should start with a code cell
* The response should contain one code cell
* Use regexes to check if interactive metadata is set correctly [jlewi/foyle#157](https://github.com/jlewi/foyle/issues/157)
  * interactive should be false unless the command matches a regex for an interactive command e.g. "kubectl.*exec.*", "docker.*run.*" etc...
* Ensure the AI doesn't generate any cells for empty input

## Reference

[Your AI Product Needs Evals](https://hamel.dev/blog/posts/evals/#level-1-unit-tests) Blog post describing the Level 1, 2, and 3 evals.

