---
title: "Honeycomb"
description: "How to use Foyle to assist you with Honeycomb."
weight: 8
---

# Why Use Honeycomb with Foyle

[Honeycomb](https://honeycomb.io) is often used to query and visualize observability data.

Using Foyle and RunMe you can

* Use AI to turn a high level intent into an actual Honeycomb query
* Capture that query as part of a playbook or incident report
* Create documents that capture the specific queries you are interested in rather than rely
  on static dashboards

<video controls style="max-width: 50%; height: auto;">
  <source src="https://storage.googleapis.com/foyle-public/videos/foylehoneycomb.mp4" type="video/mp4">
  Your browser does not support the video tag.
</video>

## How to integrate Honeycomb with Runne

This section explains how to integrate executing Honeycomb queries from Runme.

Honeycomb supports embedding queries directly in the 
[URL](https://docs.honeycomb.io/investigate/collaborate/share-query/). You can use this feature to define queries
in your notebook and then generate a URL that can be opened in the browser to view the results.

Honeycomb [queries](https://docs.honeycomb.io/api/tag/Querieshttps://docs.honeycomb.io/api/tag/Queries) are defined
in JSON. A suggested pattern to define queries in your notebook are

1. Create a code cell which uses contains the query to be executed
   * To make it readable the recommended pattern is to treat it as a multi-line JSON
     string and write it to a file
2. Use the CLI [hccli](https://github.com/jlewi/hccli) to generate the URL and open it in the browser
   * While the CLI allows passing the query as an argument this isn't nearly as human
     readable as writing it and reading it from a temporary file
Here's an example code cell

```bash
cat << EOF > /tmp/query3.json
{
  "time_range": 604800,
  "granularity": 0,  
  "calculations": [
      {
          "op": "AVG",
            "column": "prediction.ok"
      }
  ],
  "filters": [
      {
          "column": "model.name",
          "op": "=",
          "value": "llama3"
      }
  ],
  "filter_combination": "AND",
  "havings": [],
  "limit": 1000
}
EOF
hccli querytourl --query-file=/tmp/query3.json --base-url=https://ui.honeycomb.io/YOURORG/environments/production --dataset=YOURDATASET --open=true
```

* This is a simple query to get the average of the `prediction.ok` column for the `llama3` model for the last week
* Be sure to replace `YOURORG` and `YOURDATASET` with the appropriate values for your organization and dataset
  * You can determine `base-url` just by opening up any existing query in Honeycomb and copying the URL
* When you execute the cell, it will print the query and open it in your browser
* You can use the [share query](https://docs.honeycomb.io/investigate/collaborate/share-query/) feature to encode the query in the URL

## Training Foyle to be your Honeycomb Expert

To train Foyle to be your Honeycomb you follow these steps

1. Create a markdown cell expressing the high level intent you want to accomplish
1. Create a code cell which uses `hccli` to generate the query link
1. Execute the code cell
1. Foyle automatically learns how to predict every successfully executed code cell

If you need help boostrapping some initial Honeycomb JSON queries you can use [Honeycomb's Share](https://docs.honeycomb.io/investigate/collaborate/share-query/)
feature to generate the JSON from a query constructed in the UI.

### Producing Reports

RunMe's [AutoSave Feature](https://docs.runme.dev/configuration/auto-save) creates a markdown document that
contains the output of the code cells. This is great for producing a report or writing a postmortem.
When using this feature with Honeycomb you'll want to capture the output of the query. There are a couple ways
to do this

#### Permalinks
To generate permalinks, you just need to use can use `start_time` and
`end_time` to specify a fixed time range for your queries (see [Honeycomb query API](https://docs.honeycomb.io/api/tag/Queries/#operation/createQuery)). Since the 
[hccli](https://github.com/jlewi/hccli) prints the output URL it will be saved in the session outputs generated
by RunMe. You could also copy and past the URL into a markdown cell.

#### Grabbing a Screenshot

Unfortunately a Honeycomb enterprise plan is required to access query results and graphs via API. 
As a workaround, [hccli](https://github.com/jlewi/hccli) supports grabbing a screenshot of the query results using 
browser automation.

You can do this as follows

1. Restart chrome with a debugging port open 

   ```bash
   chrome --remote-debugging-port=9222
   ```

1. Login into Honeycomb

1. Add the `--out-file=/path/to/file.png` to `hccli` to specify a file to save it to

   ```
   hccli querytourl --query-file=/tmp/query3.json --base-url=https://ui.honeycomb.io/YOURORG/environments/production --dataset=YOURDATASET --open=true --out-file=/path/to/file.png
   ```
   
**Warning** Unfortunately this tends to be a bit brittle for the following reasons

* You need to restart chrome with remote debugging enabled

* `hccli` will use the most recent browser window so you need to make sure your most recent browser window
    is the one with your Honeycomb credentials. This may not be the case if you have different accounts logged
    into different chrome sessions

# Reference

* [Honeycomb template links](https://docs.honeycomb.io/investigate/collaborate/share-query/)