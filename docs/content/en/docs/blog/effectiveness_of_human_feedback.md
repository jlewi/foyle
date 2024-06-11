---
title: The Unreasonable Effectiveness of Human Feedback
linkTitle: The Unreasonable Effectiveness of Human Feedback
date: 2024-06-11
author: "[Jeremy Lewi](https://lewi.us/about)"
type: blog
description: This post presents quantitative results showing how human feedback allows Foyle to assist with building and operating Foyle. In 79% of cases, Foyle provided the correct answer, whereas ChatGPT alone would lack sufficient context to achieve the intent. Furthermore, the LLM API calls cost less than $.002 per intent whereas a recursive, agentic approach could easily cost $2-$10.
images: 
  - "/docs/blog/foyle_learning_interactions.svg"
---

Agents! Agents! Agents! Everywhere we look we are bombarded with the promises of fully autonomous agents. 
These pesky humans aren’t merely inconveniences, they are budgetary line items to be optimized away. 
All this hype leaves me wondering; have we forgotten that GPT was fine-tuned using data produced by a small army of[ human labelers](https://scale.com/blog/how-to-label-1m-data-points-week)?  Not to mention who do we think produced the 10 trillion words that foundation models are being trained on? While fully autonomous software agents are capturing the limelight on social media, systems that turn user interactions into training data like [Didact](https://research.google/blog/large-sequence-models-for-software-development-activities/), [Dosu](https://blog.langchain.dev/dosu-langsmith-no-prompt-eng/) and [Replit code repair](https://blog.replit.com/code-repair) are deployed and solving real toil.

[Foyle](https://foyle.io/) takes a user-centered approach to building an AI to help developers deploy and operate their software. 
The key premise of Foyle is to instrument a developer's workflow so that we can monitor how they turn intent into actions. 
Foyle uses that interaction data to constantly improve. A previous post described how Foyle uses this data to learn. 
This post presents quantitative results showing how feedback allows Foyle to assist with building and operating Foyle. 
In 79% of cases, Foyle provided the correct answer, whereas ChatGPT alone would lack sufficient context to achieve the intent. In particular, the results show how Foyle lets users express intent at a higher level of abstraction.

As a thought experiment, we can compare Foyle against an agentic approach that achieves the same accuracy by recursively invoking an LLM on 
Foyle’s (& [RunMe's](https://runme.dev/)) 65K lines of code but without the benefit of learning from user interactions. In this case, we estimate that Foyle could easily save between $2-$10 on LLM API calls per intent. In practice, this likely means learning from prior interactions is critical to making an affordable AI.

## Mapping Intent Into Action

The pain of deploying and operating software was famously captured in a 2010 Meme at [Google “I just want to serve 5 Tb”](https://news.ycombinator.com/item?id=29082014). The meme captured that simple objectives (e.g. serving some data) can turn into a bewildering complicated tree of operations due to system complexity and business requirements. The goal of Foyle is to solve this problem of translating intent into actions.

Since we are using Foyle to build Foyle, we can evaluate it by how well it learns to assist us with everyday tasks. 
The video below illustrates how we use Foyle to troubleshoot the AI we are building by fetching traces.

<!-- cc_load_policy turns on captions by default-->
<iframe width="560" height="315" src="https://www.youtube.com/embed/V4YcQLXosnc?cc_load_policy=1&si=-C1biEMtZhdWPspr" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>

The diagram below illustrates how Foyle works.

![Foyle Interaction Diagram](../foyle_learning_interactions.svg)


In the video, we are using Foyle to fetch the trace for a specific prediction. This is a fundamental step in any AI Engineer’s workflow. The trace contains the information needed to understand the AI’s answer; e.g. the prompts to the LLMs, the result of post-processing etc… Foyle takes the markdown produced by ChatGPT and turns it into a set of blocks and assigns each block a unique ID. So to understand why a particular block was generated we might ask for the block trace as follows


```
show the block logs for block 01HZ3K97HMF590J823F10RJZ4T
```


The first time we ask Foyle to help us it has no prior interactions to learn from so it largely passes along the request to ChatGPT and we get the following response


```
blockchain-cli show-block-logs 01HZ3K97HMF590J823F10RJZ4T
```


Unsurprisingly, this is completely wrong because ChatGPT has no knowledge of Foyle; its just guessing. The first time we ask for a trace, we would fix the command to use Foyle’s REST endpoint to fetch the logs


```
curl http://localhost:8080/api/blocklogs/01HZ3K97HMF590J823F10RJZ4T | jq . 
```


Since Foyle is instrumented to log user interactions it learns from this interaction. So the next time we ask for a trace e.g.




```
get the log for block 01HZ0N1ZZ8NJ7PSRYB6WEMH08M
```


Foyle responds with the correct answer


```
curl http://localhost:8080/api/blocklogs/01HZ0N1ZZ8NJ7PSRYB6WEMH08M | jq . 
```


Notably, this example illustrates that Foyle is learning how to map higher level concepts (e.g. block logs) into low level concrete actions (e.g. curl).


## Results

To measure Foyle’s ability to learn and assist with mapping intent into action, we created an evaluation dataset of [24 examples](https://github.com/jlewi/foyle/tree/main/data/eval) of intents specific to building and operating Foyle. The dataset consists of the following



* Evaluation Data: 24 pairs of (intent, action) where the action is a command that correctly achieves the intent
* Training Data: 27 pairs of (intent, action) representing user interactions logged by Foyle
    * These were the result of our daily use of Foyle to build Foyle

To evaluate the effectiveness of human feedback we compared using GPT3.5 without examples to GPT3.5 with examples. 
Using examples, we prompt GPT3.5 with similar examples from prior usage(the [prompt is here](https://github.com/jlewi/foyle/blob/main/app/pkg/agent/prompt.tmpl)). 
Prior examples are selected by using similarity search to find the intents most similar to the current one. To measure the correctness of the generated commands we use a version of edit distance that measures the number of arguments that need to be changed. The binary itself counts as an argument. This metric can be normalized so that 0 means the predicted command is an exact match and 1 means the predicted command is completely different (precise details are [here](https://github.com/jlewi/foyle/blob/main/tech_notes/tn003_learning_eval.md#evaluating-correctness)).

The [Table 1.](#table1) below shows that Foyle performs significantly better when using prior examples.
The full results are in the [appendix](#table2).
Notably, in 15 of the examples where using ChatGPT without examples was wrong it was completely wrong. 
This isn’t at all surprising given GPT3.5 is missing critical information to answer these questions.

<table>
  <tr>
   <td>
   </td>
   <td>Number of Examples
   </td>
   <td>Percentage
   </td>
  </tr>
  <tr>
   <td>Performed Better With Examples
   </td>
   <td>19
   </td>
   <td>79%
   </td>
  </tr>
  <tr>
   <td>Did Better or Just As Good With Examples 
   </td>
   <td>22
   </td>
   <td>91%
   </td>
  </tr>
  <tr>
   <td>Did Worse With Examples
   </td>
   <td>2
   </td>
   <td>8%
   </td>
  </tr>
</table>

<a name="table1">Table 1</a>: Shows that for 19 of the examples (79%); the AI performed better when learning from prior examples. In 22 of the 24 examples (91%) using the prior examples the AI did no worse than baseline. In 2 cases, using prior examples decreased the AI’s performance. The full results are provided in the table below.

### Distance Metric

Our distance metrics assumes there are specific tools that should be used to accomplish a task even when different solutions might produce identical answers. In the context of devops this is desirable because there is a cost to supporting a tool; e.g. ensuring it is available on all machines. As a result, platform teams are often opinionated about how things should be done.

For example to fetch the block logs for block 01HZ0N1ZZ8NJ7PSRYB6WEMH08M we measure the distance to the command


```
curl http://localhost:8080/api/blocklogs/01HZ0N1ZZ8NJ7PSRYB6WEMH08M | jq .
```


Using our metric if the AI answered


```
wget -q -O - http://localhost:8080/api/blocklogs/01HZ0N1ZZ8NJ7PSRYB6WEMH08M | yq .
```


The distance would end up being .625. The longest command consists of 8 arguments (including the binaries and the pipe operator). 3 deletions and 2 substitutions are needed to transform the actual into the expected answer which yields a distance of  ⅝=.625. So in this case, we’d conclude the AI’s answer was largely wrong even though wget produces the exact same output as curl in this case. If an organization is standardizing on curl over wget then the evaluation metric is capturing that preference.


## How much is good data worth?

A lot of agents appear to be pursuing a solution based on throwing lots of data and lots of compute at the problem. 
For example, to figure out how to “Get the log for block XYZ”, an agent could in principle crawl the 
[Foyle](https://github.com/jlewi/foyle) and [RunMe](https://github.com/stateful/runme) repositories to understand what a block is and that Foyle 
exposes a REST server to make them accessible.  That approach might cost $2-$10 in LLM calls whereas with Foyle it's less than $.002.

The Foyle repository is ~400K characters of Go Code; the RunMe Go code base is ~1.5M characters. So lets say 2M characters which is about 500K-1M tokens. 
With [GPT-4-turbo that’s ~$2-$10](https://openai.com/api/pricing/); or about 1-7 SWE minutes (assuming $90 per hour). 
If the Agent needs to call GPT4 multiple times those costs are going to add up pretty quickly.

## Where is Foyle Going

Today, Foyle is only learning single step workflows. While this is valuable, a lot of a toil involves multi step workflows. 
We’d like to extend Foyle to support this use case. This likely requires changes to how Foyle learns and how we evaluate Foyle.

Foyle only works if we log user interactions. This means we need to create a UX that is compelling enough for developers to want to use. Foyle is now integrated with [Runme](https://runme.dev/). We want to work with the Runme team to create features (e.g. [Renderers](https://github.com/stateful/vscode-runme/blob/main/README.md), [multiple executor support](https://github.com/stateful/runme/issues/593)) that give users a reason to adopt a new tool even without AI.


## How You Can Help

If you’re rethinking how you do playbooks and want to create AI assisted executable playbooks please get in touch via email [jeremy@lewi.us](mailto:jeremy@lewi.us) or by starting a discussion in [GitHub](https://github.com/jlewi/foyle/discussions). In particular, if you’re struggling with observability and want to use AI to assist in query creation and create rich artifacts combining markdown, commands, and rich visualizations, we’d love to learn more about your use case.

## Appendix: Full Results

The table below provides the prompts, RAG results, and distances for the entire evaluation dataset.

<table>
  <tr>
   <td>prompt
   </td>
   <td>best_rag
   </td>
   <td>Baseline Normalized
   </td>
   <td>Learned Distance
   </td>
  </tr>
  <tr>
   <td>Get the ids of the execution traces for block 01HZ0W9X2XF914XMG6REX1WVWG
   </td>
   <td>get the ids of the execution traces for block 01HZ3K97HMF590J823F10RJZ4T
   </td>
   <td><p style="text-align: right">
0.6666667</p>

   </td>
   <td><p style="text-align: right">
0</p>

   </td>
  </tr>
  <tr>
   <td>Fetch the replicate API token
   </td>
   <td>Show the replicate key
   </td>
   <td><p style="text-align: right">
1</p>

   </td>
   <td><p style="text-align: right">
0</p>

   </td>
  </tr>
  <tr>
   <td>List the GCB jobs that build image backend/caribou
   </td>
   <td>list the GCB builds for commit 48434d2
   </td>
   <td><p style="text-align: right">
0.5714286</p>

   </td>
   <td><p style="text-align: right">
0.2857143</p>

   </td>
  </tr>
  <tr>
   <td>Get the log for block 01HZ0N1ZZ8NJ7PSRYB6WEMH08M
   </td>
   <td>show the blocklogs for block 01HZ3K97HMF590J823F10RJZ4T
<p>
...
   </td>
   <td><p style="text-align: right">
1</p>

   </td>
   <td><p style="text-align: right">
0</p>

   </td>
  </tr>
  <tr>
   <td>How big is foyle's evaluation data set?
   </td>
   <td>Print the size of foyle's evaluation dataset
   </td>
   <td><p style="text-align: right">
1</p>

   </td>
   <td><p style="text-align: right">
0</p>

   </td>
  </tr>
  <tr>
   <td>List the most recent image builds
   </td>
   <td>List the builds
   </td>
   <td><p style="text-align: right">
1</p>

   </td>
   <td><p style="text-align: right">
0.5714286</p>

   </td>
  </tr>
  <tr>
   <td>Run foyle training
   </td>
   <td>Run foyle training
   </td>
   <td><p style="text-align: right">
0.6666667</p>

   </td>
   <td><p style="text-align: right">
0.6</p>

   </td>
  </tr>
  <tr>
   <td>Show any drift in the dev infrastructure
   </td>
   <td>show a diff of the dev infra
   </td>
   <td><p style="text-align: right">
1</p>

   </td>
   <td><p style="text-align: right">
0.4</p>

   </td>
  </tr>
  <tr>
   <td>List images
   </td>
   <td>List the builds
   </td>
   <td><p style="text-align: right">
0.75</p>

   </td>
   <td><p style="text-align: right">
0.75</p>

   </td>
  </tr>
  <tr>
   <td>Get the cloud build jobs for commit abc1234
   </td>
   <td>list the GCB builds for commit 48434d2
   </td>
   <td><p style="text-align: right">
0.625</p>

   </td>
   <td><p style="text-align: right">
0.14285715</p>

   </td>
  </tr>
  <tr>
   <td>Push the honeycomb nl to query model to replicate
   </td>
   <td>Push the model honeycomb to the jlewi repository
   </td>
   <td><p style="text-align: right">
1</p>

   </td>
   <td><p style="text-align: right">
0.33333334</p>

   </td>
  </tr>
  <tr>
   <td>Sync the dev infra
   </td>
   <td>show a diff of the dev infra
   </td>
   <td><p style="text-align: right">
1</p>

   </td>
   <td><p style="text-align: right">
0.5833333</p>

   </td>
  </tr>
  <tr>
   <td>Get the trace that generated block 01HZ0W9X2XF914XMG6REX1WVWG
   </td>
   <td>get the ids of the execution traces for block 01HZ3K97HMF590J823F10RJZ4T
   </td>
   <td><p style="text-align: right">
1</p>

   </td>
   <td><p style="text-align: right">
0</p>

   </td>
  </tr>
  <tr>
   <td>How many characters are in the foyle codebase?
   </td>
   <td>Print the size of foyle's evaluation dataset
   </td>
   <td><p style="text-align: right">
1</p>

   </td>
   <td><p style="text-align: right">
0.875</p>

   </td>
  </tr>
  <tr>
   <td>Add the tag 6f19eac45ccb88cc176776ea79411f834a12a575 to the image ghcr.io/jlewi/vscode-web-assets:v20240403t185418
   </td>
   <td>add the tag v0-2-0 to the image ghcr.io/vscode/someimage:v20240403t185418
   </td>
   <td><p style="text-align: right">
0.5</p>

   </td>
   <td><p style="text-align: right">
0</p>

   </td>
  </tr>
  <tr>
   <td>Get the logs for building the image carabou
   </td>
   <td>List the builds
   </td>
   <td><p style="text-align: right">
1</p>

   </td>
   <td><p style="text-align: right">
0.875</p>

   </td>
  </tr>
  <tr>
   <td>Create a PR description
   </td>
   <td>show a diff of the dev infra
   </td>
   <td><p style="text-align: right">
1</p>

   </td>
   <td><p style="text-align: right">
1</p>

   </td>
  </tr>
  <tr>
   <td>Describe the dev cluster?
   </td>
   <td>show the dev cluster
   </td>
   <td><p style="text-align: right">
1</p>

   </td>
   <td><p style="text-align: right">
0</p>

   </td>
  </tr>
  <tr>
   <td>Start foyle
   </td>
   <td>Run foyle
   </td>
   <td><p style="text-align: right">
1</p>

   </td>
   <td><p style="text-align: right">
0</p>

   </td>
  </tr>
  <tr>
   <td>Check for preemptible A100 quota in us-central1
   </td>
   <td>show a diff of the dev infra
   </td>
   <td><p style="text-align: right">
0.16666667</p>

   </td>
   <td><p style="text-align: right">
0.71428573</p>

   </td>
  </tr>
  <tr>
   <td>Generate a honeycomb query to count the number of traces for the last 7 days broken down by region in the foyle dataset
   </td>
   <td>Generate a honeycomb query to get number of errors per day for the last 28 days
   </td>
   <td><p style="text-align: right">
0.68421054</p>

   </td>
   <td><p style="text-align: right">
0.8235294</p>

   </td>
  </tr>
  <tr>
   <td>Dump the istio routes for the pod jupyter in namespace kubeflow
   </td>
   <td>list the istio ingress routes for the pod foo in namespace bar
   </td>
   <td><p style="text-align: right">
0.5</p>

   </td>
   <td><p style="text-align: right">
0</p>

   </td>
  </tr>
  <tr>
   <td>Sync the manifests to the dev cluster
   </td>
   <td>Use gitops to aply the latest manifests to the dev cluster
   </td>
   <td><p style="text-align: right">
1</p>

   </td>
   <td><p style="text-align: right">
0</p>

   </td>
  </tr>
  <tr>
   <td>Check the runme logs for an execution for the block 01HYZXS2Q5XYX7P3PT1KH5Q881
   </td>
   <td>get the ids of the execution traces for block 01HZ3K97HMF590J823F10RJZ4T
   </td>
   <td><p style="text-align: right">
1</p>

   </td>
   <td><p style="text-align: right">
1</p>

   </td>
  </tr>
</table>


<a name="table2">Table 2</a>. The full results for the evaluation dataset. The left column shows the evaluation prompt. The second column shows the most similar prior example (only the query is shown).  The third column is the normalized distance for the baseline AI. The 4th column is the normalized distance when learning from prior examples.
