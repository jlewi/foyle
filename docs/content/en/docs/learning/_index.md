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

Inside VSCode open an output window and select the **RunMe** output channel. 
Scroll to the top of the messages and look for a `Logger initialized` message like the one below

```
[2024-05-28T22:12:20.681Z] INFO Runme(RunmeServer): {"level":"info","ts":1716934340.6789708,"caller":"cmd/common.go:190","msg":"Logger initialized","devMode":false,"aiLogs":true,"aiLogFile":"/Users/jlewi/Library/Application Support/runme/logs/logs.2024-05-28T15:12:20.json"}
```

The field `aiLogs` will contain the file that the current instance of RunMe is using for the JSON logs.

By default, on MacOs RunMe will use the directory

```bash
/Users/${USER}/Library/Application Support/runme/logs/
```

## Configuring Learning

If you are using [RunMe.dev](https://runme.dev/) as the frontend for Foyle
then you need to configure Foyle with the location of the **directory** of RunMe's logs.
Refer to the previous section for instructions on how to locate the file
where `RunMe` is writing the logs. Then remove the filename to get the directory
where the logs are being written.

Once you know the directory run the following command to configure Foyle to use that
directory

```bash
foyle config set learner.logDirs=${RUNME_LOGS_DIR}
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
