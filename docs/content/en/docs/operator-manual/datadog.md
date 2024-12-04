---
title: "Datadog"
description: "Use Datadog to monitor Foyle"
weight: 7
---

## What You'll Learn

How to use Datadog to monitor Foyle.

## Datadog Logging


If you are running Foyle in a Kubernetes cluster with the Datadog agent then you can
do the following to have your logs show up in Datadog.

Add Datadog's universal tagging labels to the Foyle statefulset and the pod template e.g


```
labels:
    tags.datadoghq.com/env: staging
    tags.datadoghq.com/service: foyle
    tags.datadoghq.com/version: "0.1"

```

Modify your logging configuration to tell Foyle to use the "level" field to store the logging
level. Ensure that `json` is set to `true`.

```
logging:
    level: info
    ...
    # Specify the names of the fields to match what Datadog uses
    # https://docs.datadoghq.com/logs/log_configuration/parsing/?tab=matchers    
    logFields:
        level: level

    sinks:
    - json: true
```

## Prometheus (OpenMetrics)

To configure Datadog to scrape Foyle's OpenMetrics (Prometheus endpoint) add the annotation 
for the OpenMetrics check to the annotations of the pod template spec

```yaml
apiVersion: apps/v1
kind: StatefulSet
spec:
  ...
  template:
    metadata:
      annotations:
        ad.datadoghq.com/foyle.checks: |
            {
              "openmetrics": {
                "init_config": {},
                "instances": [
                  {
                    "openmetrics_endpoint": "http://%%host%%:%%port%%/metrics",
                    "namespace": "foyle",
                    "metrics": [".*"]
                  }
                ]
              }
            }           
```