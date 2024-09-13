---
description: How to monitor the quality of AI outputs
title: Monitoring AI Quality
weight: 3
---

## What You'll Learn

* How to observe the AI to understand why it generated the answers it did

## What was the actual prompt and response?

A good place to start when trying to understand the AI's responses
is to look at the actual prompt and response from the LLM that produced the cell.

You can fetch the request and response as follows

1. Get the log for a given cell
1. From the cell get the traceId of the AI generation request

```bash
CELLID=01J7KQPBYCT9VM2KFBY48JC7J0
export TRACEID=$(curl -s -X POST http://localhost:8877/api/foyle.logs.LogsService/GetBlockLog -H "Content-Type: application/json" -d "{\"id\": \"${CELLID}\"}" | jq -r .blockLog.genTraceId)
echo TRACEID=$TRACEID
```

* Given the traceId, you can fetch the request and response from the LOGS

```bash {"id":"01J7MG0CV3N8D4678XFSHTB1H7"}
curl -s -o /tmp/response.json -X POST http://localhost:8877/api/foyle.logs.LogsService/GetLLMLogs -H "Content-Type: application/json" -d "{\"traceId\": \"${TRACEID}\"}"
CODE="$?"
if [ $CODE -ne 0 ]; then
  echo "Error occurred while fetching LLM logs"
  exit $CODE
fi

```

* You can view an HTML rendering of the prompt and response
* If you disable interactive mode for the cell then vscode will render the HTML respnse inline
* **Note** There appears to be a bug right now in the HTML rendering causing a bunch of newlines to be introduced relative to what's in the actual markdown in the JSON request

```bash {"id":"01J7MM8TNZ2T1W6HN6BHJ2RN4C","interactive":"false"}
jq '.requestHtml' /tmp/response.json
```

* To view the response

```bash {"id":"01J7MMCPDJHR3T4QER1G6ANCJD","interactive":"false"}
jq '.responseHtml' /tmp/response.json
```

* To view the JSON versions of the actual requests and response 

```bash {"interactive":"false"}
jq -r '.requestJson' /tmp/response.json | jq .
```

* You can print the raw markdown of the prompt as follows 

```bash
echo $(jq -r '.requestJson' /tmp/response.json | jq '.messages[0].content[0].text')
```

```bash {"id":"01J7MMNZQZXC773MG0ARV6AB6Z"}
jq -r '.responseJson' /tmp/response.json | jq .
```