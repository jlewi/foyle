---
title: "Configuring RAG"
description: "Configuring Learning In Foyle"
weight: 1
---

## What You'll Learn

How to configure Learning in Foyle to continually learn from human feedback

## How It Works

* As you use Foyle, the AI builds a dataset of examples (input, output)
* The input is a notebook at some point in time , `t`
* The output is one more or cells that were then added to the notebook at time `t+1`
* Foyle uses these examples to get better at suggesting cells to insert into the notebook

## Configuring RAG

Foyle uses RAG to improve its predictions using its existing dataset of examples. You can control
the number of RAG results used by Foyle by setting `agent.rag.maxResults`.

```
foyle config set agent.rag.maxResults=3
```

## Disabling RAG

RAG is enabled by default. To disable it run

```
foyle config set agent.rag.enabled=false
```

To check the status of RAG get the current configuration

```
foyle config get
```

## Sharing Learned Examples

In a team setting, you should build a shared AI that learns from the feedback of all team members and assists
all members. To do this you can configure Foyle to write and read examples from a shared location like GCS.
If you'd like S3 support please vote up [issue #153](https://github.com/jlewi/foyle/issues/153).

To configure Foyle to use a shared location for learned examples

1. Create a GCS bucket to store the learned examples

   ```bash
    gsutil mb gs://my-foyle-examples
   ```

1. Configure Foyle to use the GCS bucket

   ```bash
   foyle config set learner.exampleDirs=gs://${YOUR_BUCKET}
   ```

Optionally you can configure Foyle to use a local location as well if you want to be able to use the AI without
an internet connection.

```bash
foyle config set learner.exampleDirs=gs://${YOUR_BUCKET},/local/training/examples
```
