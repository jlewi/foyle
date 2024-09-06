---
description: How to monitor Foyle with Honeycomb
title: Monitoring Foyle With Honeycomb
weight: 8
---

## What You'll Learn

* How to use Opentelemetry and Honeycomb To Monitor Foyle

## Setup

### Configure Foyle To Use Honeycomb

```sh
foyle config set telemetry.honeycomb.apiKeyFile = /path/to/apikey
```

### Download Honeycomb CLI

* [hccli](https://github.com/jlewi/hccli) is an **unoffical** CLI for Honeycomb.
* It is being developed to support using Honeycomb with Foyle.

```sh
TAG=$(curl -s https://api.github.com/repos/jlewi/hccli/releases/latest | jq -r '.tag_name')
# Remove the leading v because its not part of the binary name
TAGNOV=${TAG#v}
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
echo latest tag is $TAG
echo OS is $OS
echo Arch is $ARCH
LINK=https://github.com/jlewi/hccli/releases/download/${TAG}/hccli_${TAGNOV}_${OS}_${ARCH}
echo Downloading $LINK
wget $LINK -O /tmp/hccli
```

Move the `hccli` binary to a directory in your PATH.

```bash
chmod a+rx /tmp/hccli
sudo mv /tmp/hccli /tmp/hccli
```

On Darwin set the execute permission on the binary.

```bash
sudo xattr -d com.apple.quarantine /usr/local/bin/hccli
```

## Configure Honeycomb 

* In order for `hccli` to generate links to the Honeycomb UI it needs to know base URL of your environment.
* You can get this by looking at the URL in your browser when you are logged into Honeycomb.
* It is typically somethling like

```
https://ui.honeycomb.io/${TEAM}/environments/${ENVIRONMENT}
```

* You can set the base URL in your config

```bash
hccli config set baseURL=https://ui.honeycomb.io/${TEAM}/environments/${ENVIRONMENT}/
```

* You can check your configuration by running the get command

```bash
hccli config get
```

## Measure Acceptance Rate

* To measure the utility of Foyle we can look at how often Foyle suggestions are accepted
* When a suggestion is accepted we send a LogEvent of type `ACCEPTED`
* This creates an OTEL trace with a span with name `LogEvent` and attribute `eventType == ACCEPTED`
* We can use the query below to calculate the acceptance rate

```bash
QUERY='{
  "calculations": [
    {"op": "COUNT", "alias": "Event_Count"}
  ],
  "filters": [
    {"column": "name", "op": "=", "value": "LogEvent"},
    {"column": "eventType", "op": "=", "value": "ACCEPTED"}
  ],
  "time_range": 86400,
  "order_by": [{"op": "COUNT", "order": "descending"}]
}'

hccli querytourl --dataset foyle --query "$QUERY" --open
```

## Token Count Usage

* The cost of LLMs depends on the number of input and output tokens
* You can use the query below to look at token usage

```bash
QUERY='{
  "calculations": [
    {"op": "COUNT", "alias": "LLM_Calls_Per_Cell"},
    {"op": "SUM", "column": "llm.input_tokens", "alias": "Input_Tokens_Per_Cell"},
    {"op": "SUM", "column": "llm.output_tokens", "alias": "Output_Tokens_Per_Cell"}
  ],
  "filters": [
    {"column": "name", "op": "=", "value": "Complete"}
  ],
  "breakdowns": ["trace.trace_id"],
  "time_range": 86400,
  "order_by": [{"op": "COUNT", "order": "descending"}]
}'

hccli querytourl --dataset foyle --query "$QUERY" --open
```
