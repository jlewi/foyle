---
title: "Troubleshoot Learning"
description: "How to troubleshoot and monitor learning"
weight: 2
---

## What You'll Learn

* How to ensure learning is working and monitor learning

## Check Examples

If Foyle is learning there should be example files in ${HOME}/.foyle/training

```sh
ls -la ~/.foyle/training
```

The output should include `example.binpb` files as illustrated below.

```sh
-rw-r--r--    1 jlewi  staff   9895 Aug 28 07:46 01J6CQ6N02T7J16RFEYCT8KYWP.example.binpb
```

If there aren't any then no examples have been learned.

## Trigger Learning

Foyle's learning is triggered whenever a cell is successfully executed.

Every time you switch the cell in focus, Foyle creates a new session. 
The current session ID is displayed in the lower right hand context window
in vscode.

You can use the session ID to track whether learning occured.

## Did The Session Get Created

* Check the session log

```bash
CONTEXTID=01J7S3QZMS5F742JFPWZDCTVRG
curl -s -X POST http://localhost:8877/api/foyle.logs.SessionsService/GetSession -H "Content-Type: application/json" -d "{\"contextId\": \"${CONTEXTID}\"}" | jq .
```

* If this returns not found then no log was created for this sessions and there is a problem with Log Processing
* 
* The output should include

   * LogEvent for cell execution
   * LogEvent for session end
   * FullContext containing a notebook

* Ensure the cells all have IDs 


### Did we try to create an example from any cells?

* If Foyle tries to learn from a cell it logs a message [here](https://github.com/jlewi/foyle/blob/4288e91ac805b46103d94230b32dd1bc2f957095/app/pkg/learn/learner.go#L155)
* We can query for that log as follows

```bash
jq -c 'select(.message == "Found new training example")' ${LASTLOG}
```

* If that returns nothing then we know Foyle never tried to learn from any cells
* If it returns something then we know Foyle tried to learn from a cell but it may have failed
* If there is an error processing an example it gets logged [here](https://github.com/jlewi/foyle/blob/4288e91ac805b46103d94230b32dd1bc2f957095/app/pkg/learn/learner.go#L205)
* So we can search for that error message in the logs

```bash
jq -c 'select(.level == "Failed to write example")' ${LASTLOG}
```

```bash
jq -c 'select(.level == "error" and .message == "Failed to write example")' ${LASTLOG}
```

## Ensure Block Logs are being created

* The query below checks that block logs are being created.
* If no logs are being processed than there is a problem with the block log processing.

```bash
jq -c 'select(.message == "Building block log")' ${LASTLOG}
```

# Check Prometheus counters

Check to make sure blocks are being enqueued for learner processing

```bash
curl -s http://localhost:8877/metrics | grep learner_enqueued_total 
```

* If the number is 0 please open an issue in GitHub because there is most likely a bug

Check the metrics for post processing of blocks 

```bash
curl -s http://localhost:8877/metrics | grep learner_sessions_processed
```

* The value of `learner_sessions_processed{status="learn"}` is the number of blocks that contributed to learning
* The value of `learner_sessions_processed{status="unexecuted"}` is the number of blocks that were ignored because they were not executed
