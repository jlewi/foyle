---
title: TN011 Building An Eval Dataset
description: Building an Eval Dataset
weight: 11
author: "[Jeremy Lewi](https://lewi.us/about)"
date: 2024-09-06
status: Published
---

## Objective

* Design a solution for autobuilding an evaluation dataset from logs

## TL;DR

An evaluation dataset is critical for being able to iterate quickly. Right now
cost is a major concern because Ghost Cells use a lot of tokens due to a naive
implementation. There are a number of experiments we'd like to run to see if 
we can reduce cost without significantly impacting quality.

One way to build an evaluation dataset is from logs. With the recent logging
changes, we introduced a concept of sessions. We can use sessions to build an 
evaluation dataset.

## Sessionization Pipeline

The frontend triggers a new session whenever the active text editor changes. 
When the activeEditor changes the frontend sends a LogEvent closing the current 
session and starting a new one. The sessionId is included in requests, e.g. StreamGenerate,
using the contextId parameter. Each time a stream is initiated, the initial request
should include the notebook context.

If the active editor is a code cell and it is executed the frontend will send a LogEvent
recording execution.

The streaming log processor can build up sessions containing 

* Context
* Executed cell

We could then use the context as the input and the executed cell as the ground truth data. 
We could then experiment by how well we do predicting the executed cell.

## How Do We Handle RAG

Foyle learns from feedback. Foyle adaptively builds up a database of example (input, output) pairs 
and uses them to prompt the model. The dataset of learned examples will impact the model performance
significantly. When we do evaluation we need to take into account this learning. 

Since we are building the evaluation dataset from actual notebooks/prompts there is an ordering to the 
examples. During evaluation we can replay the sessions in the same order they occured. We can then let
Foyle adaptively build up its learned examples just as it does in production. Replaying the examples
in order ensures we don't pollute evaluation by using knowledge of the furture to predict the past.

## LLM As Judge

Lots of code cells don't contain single binary invocations but small minny programs. Therefore, the similarity
metric proposed in [TN003](/tn003_learning_eval.md/) won't work. We can use LLM as judge to decide whether
the predicted and actual command are equivalent. 

