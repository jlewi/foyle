---
title: "Datadog"
description: "How to use Foyle to assist you with Datadog."
weight: 9
---

# Why Use Datadog with Foyle

[Datadog](https://www.datadoghq.com/) is often used to store and visualize observability data.

Using Foyle and RunMe you can

* Use AI to turn a high level intent into an actual Datadog query or dashboard
* Capture that query as part of a playbook or incident report
* Create documents that capture the specific queries you are interested in rather than rely
  on static dashboards


## How to integrate Datadog Logs with Runme

This section explains how to integrate Datadog Logs with Runme

Datadog supports embedding queries directly in the 
[URL](https://docs.datadoghq.com/logs/explorer/search/). You can use this feature to define log queries
in your notebook and then generate a URL that can be opened in the browser to view the results.
The [ddctl CLI](https://github.com/jlewi/ddctl) makes it easy to generate links.

1. Download the latest CLI from [ddctl's releases page](https://github.com/jlewi/ddctl/releases)

1. Configure it with the base URL for your Datadog instance

   ```
   ddctl config set baseURL=https://acme.datadoghq.com
   ```

   * You can determine the baseURL by opening up the Datadog UI a

1. Create a code cell which uses contains the query to be executed
   * To make it readable the recommended pattern is to treat it as a multi-line JSON
     string and write it to a file
2. Use the CLI [ddctl](https://github.com/jlewi/ddctl) to generate the URL and open it in the browser
   * While the CLI allows passing the query as an argument this isn't nearly as human
     readable as writing it and reading it from a temporary file
Here's an example code cell

```bash
cat << EOF > /tmp/query.yaml
query: service:foyle @contextId:01JEF30XB9A
EOF
ddctl logs querytourl --query-file=/tmp/query.yaml
```

* The value of the YAML file should be a map containing the query arguments to include in the link
* Some useful query arguments are

  * `live`: Set this to true if you always want to show the latest logs rather than fixing to a particular time range

* You can use the `ddctl` command line flags `--duration` and `--end-time` to control the time window shown

## Training Foyle to be your Datadog Expert

To train Foyle to be your Datadog expert you follow these steps

1. Create a markdown cell expressing the high level intent you want to accomplish
1. Create a code cell which uses `ddctl` to generate the logs explorer link
1. Execute the code cell
1. Foyle automatically learns how to predict every successfully executed code cell
