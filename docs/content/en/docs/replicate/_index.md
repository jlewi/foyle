---
title: "LLAMA3 on Replicate"
description: "Using LLAMA3 on Replicate with Foyle"
weight: 4
---

## What You'll Learn

How to configure Foyle to use [LLAMA3 hosted on Replicate](https://replicate.com/meta/meta-llama-3-8b-instruct?input=http)

## Prerequisites

1. You need a [Replicate account](https://replicate.com/docs)

## Setup Foyle To Use LLAMA3 on Replicate

1. Get an [API Token from Replicate](https://replicate.com/signin?next=/account/api-tokens) and save it to a file

1. Configure Foyle to use this API key

```
foyle config set replicate.apiKeyFile=/path/to/your/key/file
```
1. Configure Foyle to use [LLAMA3 hosted on Replicate](https://replicate.com/meta/meta-llama-3-8b-instruct?input=http)

1. Configure Foyle to use the appropriate model deployments

```
foyle config set  agent.model=meta/meta-llama-3-8b-instruct
foyle config set  agent.modelProvider=replicate                
```

## How It Works

Foyle uses 2 Models

* A Chat model to generate completions
* An embedding model to compute embeddings for RAG

Replicate provides hosted versions of [meta/llama-3-8b-instruct](https://replicate.com/meta/meta-llama-3-8b-instruct) 
and [meta/llama-3-8b-instruct](https://replicate.com/meta/meta-llama-3-70b-instruct) which are chat models. Notably,
these models are kept warm so when you send predictions Replicate doesn't need to boot up new instances leading to
long latencies. Replicate also provides an [OpenAI proxy](https://lifeboat.replicate.dev/) so you can use the OpenAI
APIs to generate responses.

Unfortunately, Replicate doesn't provide hosted, always versions of the embedding models. So Foyle continues to
use OpenAI for the embedding models.
