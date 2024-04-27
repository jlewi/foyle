---
title: "Learning"
description: "How to use Foyle to learn from human feedback"
weight: 2
---

## What You'll Learn

How to configure Foyle to continually learn from human feedback

## Configure Foyle to use RAG

To configure Foyle to use RAG

```
foyle config set agent.rag.enabled=true
foyle config set agent.rag.maxResults=3
```

## Learning from past mistakes

To learn from past mistakes you should periodically run the command

```
foyle logs process
foyle learn
```

When you run this command Foyle analyzes its logs to learn from implicit human feedback. 