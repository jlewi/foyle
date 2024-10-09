---
title: "Cloud Logging"
description: "Use Google Cloud Logging for Foyle logs"
weight: 7
---

## What You'll Learn

How to use Google Cloud Logging to monitor Foyle.

## Google Cloud Logging

If you are running Foyle locally but want to stream logs to Cloud Logging you can edit the logging stanza in your
config as follows:

```
logging:
  sinks:
    - json: true
      path: gcplogs:///projects/${PROJECT}/logs/foyle     
```

**Remember** to substitute in your actual project for ${PROJECT}.

In the Google Cloud Console you can find the logs using the query

```
logName = "projects/${PROJECT}/logs/foyle"
```

While Foyle logs to JSON files, Google Cloud Logging is convenient for querying and viewing your logs.

## Logging to Standard Error

If you want to log to standard error you can set the logging stanza in your config as follows:

```
logging:
  sinks:
    - json: false
      path: stderr     
```



