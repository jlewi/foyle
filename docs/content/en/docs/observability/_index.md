---
title: "Observability"
description: "Observability for Foyle"
weight: 7
---

## What You'll Learn

How to use OpenTelemetry to monitor Foyle.

## Configure Honeycomb

To configure Foyle to use Honeycomb as a backend you just need to set the `telemetry.honeycomb.apiKeyFile` 
configuration option to the path of a file containing your Honeycomb API key.

```
foyle config set telemetry.honeycomb.apiKeyFile = /path/to/apikey
```