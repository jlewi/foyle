---
title: "Use Anthropic"
description: "Using Anthropic with Foyle"
weight: 2
---

## What You'll Learn

How to configure Foyle to use [Anthropic Models](https://docs.anthropic.com/en/docs/about-claude/models)

## Prerequisites

1. You need a [Anthropic account](https://docs.anthropic.com/en/docs/quickstart)

## Setup Foyle To Use Anthropic Models

1. Get an [API Token from the Anthropic Console](https://console.anthropic.com/settings/keys) and save it to a file

1. Configure Foyle to use this API key

```
foyle config set anthropic.apiKeyFile=/path/to/your/key/file
```

1. Configure Foyle to use the desired [Antrhopic Model](https://docs.anthropic.com/en/docs/about-claude/models)

```
foyle config set  agent.model=claude-3-5-sonnet-20240620
foyle config set  agent.modelProvider=anthropic                
```

## How It Works

Foyle uses 2 Models

* A Chat model to generate completions
* An embedding model to compute embeddings for RAG

[Anthropic's models](https://docs.anthropic.com/en/docs/about-claude/models).

Anthropic doesn't provide embedding models so Foyle continues to
use OpenAI for the embedding models. At some, we may add support for [Voyage AI's embedding models](https://docs.anthropic.com/en/docs/build-with-claude/embeddings).
