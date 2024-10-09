---
title: "BigQuery"
description: "How to use Foyle to assist you with BigQuery."
weight: 7
---

# Why Use Foyle with BigQuery

If you are using [BigQuery](https://cloud.google.com/bigquery/docs/introduction) as your data warehouse then you'll want to query that data. Constructing the right query
is often a barrier to leveraging that data. Using Foyle you can train a personalized AI assistant to be an expert in answering
high level questions using your data warehouse.

<video controls style="max-width: 50%; height: auto;">
  <source src="https://storage.googleapis.com/foyle-public/videos/bigqueryfoyle.mp4" type="video/mp4">
  Your browser does not support the video tag.
</video>

## Prerequisites

Install the [Data TableRenderers](https://marketplace.visualstudio.com/items?itemName=RandomFractalsInc.vscode-data-table) extension in vscode 
* This extension can render notebook outputs as tables
* In particular, this extension can render JSON as nicely formatted, interactive tables

## How to integrate BigQuery with RunMe

Below is an example code cell illustrating the recommended pattern for executing BigQuery queries within RunMe.

```bash
cat <<EOF >/tmp/query.sql
SELECT
  DATE(created_at) AS date,
  COUNT(*) AS pr_count
FROM
  \`githubarchive.month.202406\`
WHERE
  type = "PullRequestEvent"
  AND 
  repo.name = "jlewi/foyle"
  AND
  json_value(payload, "$.action") = "closed"
  AND
  DATE(created_at) BETWEEN DATE_SUB(CURRENT_DATE(), INTERVAL 28 DAY) AND CURRENT_DATE()
GROUP BY
  date
ORDER BY
  date;
EOF

export QUERY=$(cat /tmp/query.sql)
bq query --format=json --use_legacy_sql=false "$QUERY"
```

As illustrated above, the pattern is to use `cat` to write the query to a file. This allows
us to write the query in a more human readable format. Including the entire SQL query in the code
cell is critical for enabling Foyle to learn the query.

The output is formatted as JSON. This allows the output to be rendered using the
[Data TableRenderers](https://marketplace.visualstudio.com/items?itemName=RandomFractalsInc.vscode-data-table) extension.

**Before** executing the code cell click the **configure** button in the lower right hand side of the cell
and then uncheck the box under "interactive". Running the cell in interactive mode prevents the output from
being rendered using [Data TableRenders](https://marketplace.visualstudio.com/items?itemName=RandomFractalsInc.vscode-data-table).
For more information refer to the [RunMe Cell Configuration Documentation](https://docs.runme.dev/configuration/cell-level#human-friendly-output).

We then use `bq query` to execute the query.

## Controlling Costs

BigQuery charges based on the amount of data scanned. To prevent accidentally running expensive queries you can
use the `--maximum_bytes_billed` to limit the amount of data scanned. BigQuery currently charges 
[$6.25 per TiB](https://cloud.google.com/bigquery/pricing?utm_source=google&utm_medium=cpc&utm_campaign=na-US-all-en-dr-bkws-all-all-trial-e-dr-1707554&utm_content=text-ad-none-any-DEV_c-CRE_665665924753-ADGP_Hybrid%20%7C%20BKWS%20-%20MIX%20%7C%20Txt-Data%20Analytics-BigQuery%20Pricing-KWID_43700077225652872-kwd-143282056846&utm_term=KW_bigquery%20pricing-ST_bigquery%20pricing&gad_source=1&gclid=CjwKCAjw1emzBhB8EiwAHwZZxXo8uWH1p4uadk24MVHGjFH31J9NG3GCqCWKEg2uNnf83lgsIxcdfhoC5JYQAvD_BwE&gclsrc=aw.ds).

## Troubleshooting
### Output Isn't Rendered Using Data TableRenderers

If the output isn't rendered using [Data TableRenderers](https://marketplace.visualstudio.com/items?itemName=RandomFractalsInc.vscode-data-table)
there are a few things to check

1. Click the ellipsis to the left of the the upper left hand corner and select **change presentation**
   * This should show you different mime-types and options for rendering them
   * Select **Data table** 

1. Another problem could be that `bq` is outputting status information while running the query
and this is interfering with the rendering. You can work around this by redirecting stderr to `/dev/null`. For
example,

   ```bash
      bq query --format=json --use_legacy_sql=false "$QUERY" 2>/dev/null
   ```
   
1. Try explicitly configuring the mime type by opening the [cell configuration](https://docs.runme.dev/configuration/cell-level#configuration-of-cell) and then
   1. Go to the advanced tab
   2. Entering "application/json" in the mime type field



