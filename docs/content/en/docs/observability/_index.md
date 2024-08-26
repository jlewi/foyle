---
title: "Observability"
description: "Observability for Foyle"
weight: 7
---

## What You'll Learn

How to use OpenTelemetry and Logging to monitor Foyle.

## Configure Honeycomb

To configure Foyle to use Honeycomb as a backend you just need to set the `telemetry.honeycomb.apiKeyFile` 
configuration option to the path of a file containing your Honeycomb API key.

```
foyle config set telemetry.honeycomb.apiKeyFile = /path/to/apikey
```

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



