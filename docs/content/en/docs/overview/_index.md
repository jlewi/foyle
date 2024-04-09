---
title: Overview
weight: 1
---

Foyle is a project aimed at building agents to help software developers
deploy and operate software. 

Foyle moves software operations out of the shell and into a literate
environment. Using Foyle, users 

* operate their infrastructure using VSCode Notebooks
  * Foyle executes cells containing shell commands by using a simple API to execute them locally or remotely 
    (e.g. in a container)  
* get help from AI's (ChatGPT, Claude, Bard, etc...) which add cells to the notebook

Using VSCode Notebooks makes it easy for Foyle to learn from human feedback and get smarter over time.

## What is Foyle?

Foyle is a web application that runs locally. When you run Foyle you can open up a VSCode Notebook 
in your browser. Using this notebook you can 

* Ask Foyle to assist you by adding cells to the notebook that contain markdown or commands to execute
* Execute the commands in the cells
* Ask Foyle to figure out what to do next based on the output of previous commands

## Why use Foyle?

* You are tired of copy/pasting commands from ChatGPT/Claude/Bard into your shell
* You want to move operations into a literate environment so you can easily capture the reasoning behind the commands you are executing
* You want to collect data to improve the AI's you are using to help you operate your infrastructure

## Building Better Agents with Better Data

The goal of foyle is to use this literate environment to collect two types of data
with the aim of building better agents

1. Human feedback on agent suggestions
1. Human examples of reasoning traces

### Human feedback

We are all asking AI's (ChatGPT, Claude, Bard, etc...) to write commands to perform
operations. These AI's often make mistakes. This is especially true when the correct answer depends on internal knowledge which the AI doesn't have.

Consider a simple example, lets ask ChatGPT

```
What's the gcloud logging command to fetch the logs for the hydros manifest named hydros?
```

ChatGPT responds with

```
gcloud logging read "resource.labels.manifest_name='hydros' AND logName='projects/YOUR_PROJECT_ID/logs/hydros'"
```

This is wrong; ChatGPT even suspects its likely to be wrong because it doesn't have any knowledge of the logging scheme used by [hydros](https://github.com/jlewi/hydros).
As users, we would most likely copy the command into our shell and iterate on it until we come up with the correct command; i.e

```
gcloud logging read 'jsonPayload."ManifestSync.Name"="hydros"'
```

This feedback is gold. We now have ground truth data `(prompt, human corrected answer)` that we could use to improve our AIs. Unfortunately, today's UX (copy and pasting into the shell)
means we are throwing this data away.

The goal of foyle's literate environment is to create a UX that allows us to easily capture

1. The original prompt
1. The AI provided answer
1. Any corrections the user makes.

Foyle aims to continuously use this data to retrain the AI so that it gets better the more you use it.

### Reasoning Traces

Everyone is excited about the ability to build agents that can reason and perform complex tasks e.g. [Devin](https://www.cognition-labs.com/introducing-devin).
To build these agents we will need examples of reasoning traces that can be used to train the agent. This need is especially acute
when it comes to building agents that can work with our private, internal systems.

Even when we start with the same tools (Kubernetes, GitHub Actions, Docker, Vercel, etc...), we end up adding tooling on top of that.
These can be simple scripts to encode things like naming conventions or they may be complex internal developer platforms. Either way,
agents need to be trained to understand these platforms if we want them to operate software on our behalf.

Literate environments (e.g. [Datadog Notebooks](https://docs.datadoghq.com/notebooks/)) are great for routine operations and troubleshooting.
Using literate environments to operate infrastructure leads to a self documenting process that automatically captures

1. Human thinking/analysis
1. Commands/operations executed
1. Command output

Using literate environments provides a superior experience to copy/pasting commands and outputs into a GitHub issue/slack channel/Google Doc to
create a record of what happened.

More importantly, the documents produced by literate environments contain essential information for training agents to operate our infrastructure.

<!-- ## How you can help 

TODO(jeremy): We should ask people try it out but first we need it to be working enough 
for people to try out.
-->

## Where should I go next?
* [Getting Started](/docs/getting-started/): Get started with $project
* [FAQ](/docs/reference/faq/): Frequently asked questions
