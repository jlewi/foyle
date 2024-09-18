---
author: '[Jeremy Lewi](https://lewi.us/about)'
date: 2024-09-17T00:00:00Z
description: Why Does Foyle Cost So Much
status: Published
title: TN012 Why Does Foyle Cost So Much
weight: 11
---

## Objective

* Understand Why Foyle Costs So Much

## TL;DR

Since enabling Ghost Cells, Foyle has been using a lot of tokens.

We need to understand why and see if we can reduce the cost without significantly impacting quality. In particular,
there are two different drivers of our input tokens

* How many completions we generate for a particular context (e.g. as a user types)
* How much context we include in each completion

## Background

### What is a session

The basic interaction in Foyle is as follows:

1. User selects a cell to begin editing
   * In VSCode each cell is a text editor so this corresponds to changing the active text editor

1. A user starts editing the cell

1. As the user edits the cell (e.g. types text), Foyle generates suggested cells to insert based on the current context of the cell.
   * The suggested cells are inserted as a Ghost Cell

1. User switches to a different cell

We refer to the sequence of events as above as a session. Each session is associated with a cell that the user is editing.
The session starts when the user activates that cell and the session ends when the user switches to a different cell.

## How Much Does Claude Cost

The graph below shows my usage of Claude over the past month.

![Claude Usage](claude_usage.png)
Figure 1. My Claude usage over the past month. The usage is predominantly from Foyle although there some other minor usages (e.g. with Continue.dev). This is for Sonnet 3.5 which is \$3 per 1M input tokens and \$15 per 1M output tokens.

Figure 1. Shows my Claude usage. The cost breakdown is as follows.

| Type | # Tokens | Cost |
|------|----------|------|
| Input | 16,209,097 | $48.6 |
| Output | 337,018| $4.5 |

So cost is dominated by input tokens.

I think there are on the order of 500 to 1000 sesions in this time frame. This means each session is about \$.05 to \$.10 per session.

We can use logs to count the number of events in a [14 day window](https://cloudlogging.app.goo.gl/WtF3TXKtEuRkjJji9).

| Event Type | Count |
|------------|-------|
|Session Start| 792 |
|Agent.StreamGenerate| 332 |
|Agent.Generate | 1487 |


I'm surprised that the number of Session Start events is higher than the number of Agent.StreamEvents. I would have expected `Agent.StreamGenerate` to be higher than `Session Start` since for a given session the stream might be recreated multiple times (e.g. due to timeouts). A SessionStart is reported when the [active text editor changes](https://github.com/stateful/vscode-runme/blob/1a48894c9fcada0234a5695b7ec3ed7b7fb803c6/src/extension/ai/ghost.ts#L220). A stream should be created when the [cell contents change](https://github.com/stateful/vscode-runme/blob/1a48894c9fcada0234a5695b7ec3ed7b7fb803c6/src/extension/ai/ghost.ts#L315). So the data implies that we switch the active text editor without changing the cell contents.

The number of `Agent.Generate` events is ~4.5x higher than `Agent.StreamGenerate`. This means on average we do at least 4.5 completions per session. However, a session could consist of multiple `StreamGenerate` requests because the stream can be interrupted and restarted; e.g. because of timeouts.

