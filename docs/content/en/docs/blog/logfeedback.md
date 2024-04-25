---
title: Logging Implicit Human Feedback
linkTitle: Logging Human Feedback
date: 2024-04-24
author: "[Jeremy Lewi](https://lewi.us/about)"
type: blog
descrption: Foyle is an open source assistant to help software developers deal with the pain of devops. One of Foyle's central premises is that creating a UX that implicitly captures human feedback is critical to building AIs that effectively assist us with operations. This post describes how Foyle logs that feedback.
images: 
  - "/docs/blog/cellids.svg"
---


[Foyle](https://foyle.io/) is an open source assistant to help software developers deal with the pain of devops. Developers are expected to operate their software which means dealing with the complexity of Cloud. Foyle aims to simplify operations with AI.  One of Foyle's central premises is that creating a UX that implicitly captures human feedback is critical to building AIs that effectively assist us with operations. This post describes how Foyle logs that feedback.

## The Problem

As software developers, we all ask AIs (ChatGPT, Claude, Bard, Ollama, etc.…) to write commands to perform operations. These AIs often make mistakes. This is especially true when the correct answer depends on internal knowledge, which the AI doesn’t have.

* What region, cluster, or namespace is used for dev vs. prod?
* What resources is the internal code name "caribou" referring to?
* What logging schema is used by our internal CICD tool?

The experience today is

* Ask an assistant for one or more commands
* Copy those commands to a terminal
* Iterate on those commands until they are correct

When it comes to building better AIs, the human feedback provided by the last step is gold. Yet today’s UX doesn’t allow us to capture this feedback easily. At best, this feedback is often collected out of band as part of a data curation step. This is problematic for two reasons. First, it's more expensive because it requires paying for labels (in time or money). Second, if we’re dealing with complex, bespoke internal systems, it can be hard to find people with the requisite expertise.

## Frontend

If we want to collect human feedback, we need to create a single unified experience for

1. Asking the AI for help
2. Editing/Executing AI suggested operations

If users are copying and pasting between two different applications the likelihood of being able to instrument it to collect feedback goes way down. Fortunately, we already have a well-adopted and familiar pattern for combining exposition, commands/code, and rich output. Its notebooks. 

Foyle’s frontend is VSCode notebooks. In Foyle, when you ask an AI for assistance, the output is rendered as cells in the notebook. The cells contain shell commands that can then be used to execute those commands either locally or remotely using the [notebook controller API](https://code.visualstudio.com/api/extension-guides/notebook#controller), which talks to a Foyle server. Here's a short video
illustrating the key interactions.

<!-- cc_load_policy turns on captions by default-->
<iframe width="560" height="315" src="https://www.youtube.com/embed/gU1XyRsV2n4?cc_load_policy=1&si=SNlYWKlgCmo4vXPi" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>


Crucially, cells are central to how Foyle creates a UX that automatically collects human feedback. When the AI generates a cell, it attaches a UUID to that cell. That UUID links the cell to a trace that captures all the processing the AI did to generate it (e.g any LLM calls, RAG calls, etc…). In VSCode, we can use cell metadata to track the UUID associated with a cell.

When a user executes a cell, the frontend sends the contents of the cell along with its UUID to the Foyle server. The UUID then links the cell to a trace of its execution. The cell’s UUID can be used to join the trace of how the AI generated the cell with a trace of what the user actually executed. By comparing the two we can easily see if the user made any corrections to what the AI suggested.

![cell ids interaction diagram](../cellids.svg)

## Traces

Capturing traces of the AI and execution are essential to logging human feedback. Foyle is designed to run on your infrastructure (whether locally or in your Cloud). Therefore, it's critical that Foyle not be too opinionated about how traces are logged. Fortunately, this is a well-solved problem. The standard pattern is:

1. Instrument the app using [structured logs](https://newrelic.com/blog/how-to-relic/structured-logging)
2. App emits logs to stdout and stderr
3. When deploying the app collect stdout and stderr and ship them to whatever backend you want to use (e.g. [Google Cloud Logging](https://cloud.google.com/logging?utm_source=google&utm_medium=cpc&utm_campaign=na-US-all-en-dr-skws-all-all-trial-e-dr-1707554&utm_content=text-ad-none-any-DEV_c-CRE_665665942555-ADGP_Hybrid+%7C+SKWS+-+MIX+%7C+Txt-Operations-Cloud+Logging-KWID_43700077212829919-kwd-35764926831&utm_term=KW_cloud+logging-ST_cloud+logging&gad_source=1&gclid=Cj0KCQjwiYOxBhC5ARIsAIvdH53Ojmkc3b--6hxt9l3zqmJfXvyrE2aTDw8fyb90g5SKrfPAbHnBajAaAtIrEALw_wcB&gclsrc=aw.ds&hl=en), [Datadog](https://www.datadoghq.com/), Splunk etc…)

When running locally, setting up an agent to collect logs can be annoying, so Foyle has the built-in ability to log to files. We are currently evaluating the need to add direct support for other backends like Cloud Logging. This should only matter when running locally because if you're deploying on Cloud chances are your infrastructure is already instrumented to collect stdout and stderr and ship them to your backend of choice.


## Don’t reinvent logging

Using existing logging libraries that support structured logging seems so obvious to me that it hardly seems worth mentioning. Except, within the AI Engineering/LLMOps community, it’s not clear to me that people are reusing existing libraries and patterns. Notably, I’m seeing a new class of observability solutions that require you to instrument your code with their SDK. I think this is undesirable as it violates the separation of concerns between how an application is instrumented and how that telemetry is stored, processed, and rendered. My current opinion is that Agent/LLM observability can often be achieved by reusing existing logging patterns. So, in defense of that view, here’s the solution I’ve opted for.

Structured logging means that each log line is a JSON record which can contain arbitrary fields. To Capture LLM or RAG requests and responses, I log [them](https://github.com/jlewi/foyle/blob/7db00af74bef5d1cc460e6499a3e7aac6c4291a0/app/pkg/agent/agent.go#L117); e.g.


```
request := openai.ChatCompletionRequest{
			Model:       a.config.GetModel(),
			Messages:    messages,
			MaxTokens:   2000,
			Temperature: temperature,
	    }

log.Info("OpenAI:CreateChatCompletion", "request", request)

```


This ends up logging the request in JSON format. Here’s an example

```
{
  "severity": "info",
  "time": 1713818994.8880482,
  "caller": "agent/agent.go:132",
  "function": "github.com/jlewi/foyle/app/pkg/agent.(*Agent).completeWithRetries",
  "message": "OpenAI:CreateChatCompletion response",
  "traceId": "36eb348d00d373e40552600565fccd03",
  "resp": {
    "id": "chatcmpl-9GutlxUSClFaksqjtOg0StpGe9mqu",
    "object": "chat.completion",
    "created": 1713818993,
    "model": "gpt-3.5-turbo-0125",
    "choices": [
      {
        "index": 0,
        "message": {
          "role": "assistant",
          "content": "To list all the images in Artifact Registry using `gcloud`, you can use the following command:\n\n```bash\ngcloud artifacts repositories list --location=LOCATION\n```\n\nReplace `LOCATION` with the location of your Artifact Registry. For example, if your Artifact Registry is in the `us-central1` location, you would run:\n\n```bash\ngcloud artifacts repositories list --location=us-central1\n```"
        },
        "finish_reason": "stop"
      }
    ],
    "usage": {
      "prompt_tokens": 329,
      "completion_tokens": 84,
      "total_tokens": 413
    },
    "system_fingerprint": "fp_c2295e73ad"
  }
}

```


A single request can generate multiple log entries. To group all the log entries related to a particular request, I attach a trace id to each log message. 


```
func (a *Agent) Generate(ctx context.Context, req *v1alpha1.GenerateRequest) (*v1alpha1.GenerateResponse, error) {
   span := trace.SpanFromContext(ctx)
   log := logs.FromContext(ctx)
   log = log.WithValues("traceId", span.SpanContext().TraceID())

```


Since I’ve instrumented [Foyle](https://foyle.io/) with open telemetry(OTEL), each request is automatically assigned a trace id. I attach that trace id to all the log entries associated with that request. Using the trace id assigned by OTEL means I can link the logs with the open telemetry trace data.

OTEL is an open standard for [distributed tracing](https://docs.honeycomb.io/get-started/basics/observability/concepts/distributed-tracing/). I find OTEL great for instrumenting my code to understand how long different parts of my code took, how often errors occur and how many requests I’m getting. You can use OTEL for LLM Observability; here's an [example](https://docs.honeycomb.io/get-started/start-building/llm/). However, I chose logs because as noted in the next section they are easier to mine.

### Aside: Structured Logging In Python

Python’s logging module supports structured logging. In Python you can use the extra argument to pass an arbitrary dictionary of values.  In python the equivalent would be:


```
logger.info("OpenAI:CreateChatCompletion", extra={'request': request, "traceId": traceId})
```


You then configure the logging module to use the [python-json-logger](https://pypi.org/project/python-json-logger/) formatter to emit logs as JSON. Here’s the [logging.conf](https://github.com/jlewi/notes/blob/main/py/pkg/logging.conf) I use for Python.


### Logs Need To Be Mined

Post-processing your logs is often critical to unlocking the most valuable insights. In the context of Foyle, I want a record for each cell that captures how it was generated and any subsequent executions of that cell. To produce this, I need to write a simple ETL pipeline that does the following:



* Build a trace by grouping log entries by trace ID
* Reykey each trace by the cell id the trace is associated with
* Group traces by cell id

This logic is highly specific to Foyle. No observability tool will support it out of box. 

Consequently, a key consideration for my observability backend is how easily it can be wired up to my preferred ETL tool. Logs processing is such a common use case that most existing logging providers likely have good support for exporting your logs. With Google Cloud Logging for example it's easy to setup log sinks to route logs to GCS, BigQuery or PubSub for additional processing.


### Visualization

The final piece is being able to easily visualize the traces to inspect what’s going on. Arguably, this is where you might expect LLM/AI focused tools might shine. Unfortunately, as the previous section illustrates, the primary way I want to view Foyle’s data is to look at the processing associated with a particular cell. This requires post-processing the raw logs. As a result, out of box visualizations won’t let me view the data in the most meaningful way. 

To solve the visualization problem, I’ve built a lightweight progressive web app(PWA) in Go ([code](https://github.com/jlewi/foyle/tree/main/app/pkg/logsviewer)) using [maxence-charriere/go-app](https://github.com/maxence-charriere/go-app). While I won’t be winning any design awards, it allows me to get the job done quickly and reuse existing libraries. For example, to render markdown as HTML I could reuse the Go libraries I was already using ([yuin/goldmark](https://github.com/yuin/goldmark)). More importantly, I don’t have to wrestle with a stack(typescript, REACT, etc…) that I’m not proficient in. With [Google Logs Analytics](https://cloud.google.com/logging/docs/analyze/query-and-view), I can query the logs using SQL. This makes it very easy to join and process a trace in the web app. This makes it possible to view traces in real-time without having to build and deploy a streaming pipeline. 


## Try Foyle

Please consider following the [getting started guide](/docs/getting-started/) to try out an early version of Foyle and share your thoughts by email([jeremy@lewi.us](mailto:jeremy@lewi.us)) on [GitHub](https://github.com/jlewi/foyle/issues)(jlewi/foyle) or on twitter ([@jeremylewi](https://twitter.com/jeremylewi))!

## About Me

I’m a Machine Learning platform engineer with over 15 years of experience. I create platforms that facilitate the rapid deployment of AI into production. I worked on Google’s Vertex AI where I created Kubeflow, one of the most popular OSS frameworks for ML. 

I’m open to new [consulting work](https://lewi.us/hire) and other forms of advisory. If you need help with your project, send me a brief email at [jeremy@lewi.us](mailto:jeremy@lewi/us).

## Acknowledgements

I really appreciate [Hamel Husain](https://hamel.dev/) and (Joseph Gleasure)[https://josephgleasure.com/] reviewing and editing this post.