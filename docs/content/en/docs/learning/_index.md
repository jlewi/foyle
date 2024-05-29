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

## Enabling Logging In RunMe

If you are using [RunMe.dev](https://runme.dev/) as the frontend for Foyle
then you need to configure RunMe to enable the AI logging experiment

1. Inside vscode open the settings panel
2. Enable the option `Runme › Experiments: Ai Logs`

Now that logging is enabled. You can verify that the logs are being written and identify the location
of the logs.

By default, on MacOs RunMe will use

```bash
/Users/${USER}/Library/Application Support/runme/logs/
```

Inside VSCode open an output window and select the **RunMe** output channel. 
Scroll to the top of the messages and look for a `Logger initialized` message like the one below

```
[2024-05-28T22:12:20.681Z] INFO Runme(RunmeServer): {"level":"info","ts":1716934340.6789708,"caller":"cmd/common.go:190","msg":"Logger initialized","devMode":false,"aiLogs":true,"aiLogFile":"/Users/jlewi/Library/Application Support/runme/logs/logs.2024-05-28T15:12:20.json"}
```

The field `aiLogs` will contain the directory where the JSON logs are written.

## Configuring Learning

If you are using [RunMe.dev](https://runme.dev/) as the frontend for Foyle
then you need to configure Foyle with the location of the logs.
Refer to the previous section for instructions on how to locate the file
where `RunMe` is writing the logs. Then remove the filename to get the directory
where the logs are being written.

Once you know the directory run the following command to configure Foyle to use that
directory

```bash
foyle config set learner.logDirs=${RUNME_LOGS_DIR}
```

## Learning from past mistakes

To learn from past mistakes you should periodically run the command

```
foyle logs process
foyle learn
```

When you run this command Foyle analyzes its logs to learn from implicit human feedback. 