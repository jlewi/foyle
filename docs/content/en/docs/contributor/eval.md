---
description: Documentation for contributors to Foyle
title: Evaluation 
weight: 10
---


## What You'll Learn

* How to setup and run experiments to evaluate the quality of the AI responses in Foyle

## Produce an evaluation dataset

In order to evaluate Foyle you need a dataset of examples that consist of notebooks and the expected cells to
be appended to the notebook. If you've been using Foyle then you can produce a dataset of examples from the logs.

```bash
DATA_DIR=<Directory where you want to store the examples>
curl -X POST http://localhost:8877/api/foyle.logs.SessionsService/DumpExamples -H "Content-Type: application/json" -d "{\"output\": \"$DATA_DIR\"}"
```
This assumes you are running Foyle on the default port of 8877. If you are running Foyle on a different port you will need 
to adjust the URL accordingly.

Everytime you execute a cell it is logged to Foyle. Foyle turns this into an example where the input is all the cells
in the notebook before the cell you executed and the output is the cell you executed. This allows us to evaluate
how well Foyle does generating the executed cell given the preceding cells in the notebook.

## Setup Foyle

Create a Foyle configuration with the parameters you want to test.

Create a directory to store the Foyle configuration.

```bash
make ${EXPERIMENTS_DIR}/${NAME}
```

Edit `{EXPERIMENTS_DIR}/${NAME}/config.yaml` to set the parameters you want to test; e.g.

* Assign a different port to the agent to avoid conflicting with other experiments or the production agent
* Configure the Model and Model Provider
* Configure RAG

## Configure the Experiment

Create the file `{$EXPERIMENTS_DIR}/${NAME}/config.yaml`

```yaml
kind: Experiment
apiVersion: foyle.io/v1alpha1
metadata:
  name: "gpt35"
spec:
  evalDir:      <PATH/TO/DATA>
  agentAddress: "http://localhost:<PORT>/api"
  outputDB:     "{$EXPERIMENTS_DIR}/${NAME}/results.sqlite"
```

* Set evalDir to the directory where you dumped the session to evaluation examples
* Set agentAddress to the address of the agent you want to evaluate
  * Use the port you assigned to the agent in `config.yaml`
* Set outputDB to the path of the sqlite database to store the results in


## Running the Experiment

Start an instance of the agent with the configuration you want to evaluate.

```bash
foyle serve --config=${EXPERIMENTS_DIR}/${NAME}/config.yaml
```

Run the experiment

```bash
foyle apply ${EXPERIMENTS_DIR}/${NAME}/experiment.yaml
```

## Analyzing the Results

You can use sqlite to query the results.

The queries below compute the following

* The number of results in the dataset
* The number of results where an error prevented a response from being generated
* The distribution of the `cellsMatchResult` field in the results
  * A value of `MATCH` indicates the generated cell matches the expected cell
  * A value of `MISMATCH` indicates the generated cell doesn't match the expected cell
  * A value of `""` (empty string) indicates no value was computed most likely because an error occurred

```bash
# Count the total number of results
sqlite3 --header --column ${RESULTSDB} "SELECT COUNT(*) FROM results"

# Count the number of errors
sqlite3 --header --column ${RESULTSDB} "SELECT COUNT(*) FROM results WHERE json_extract(proto_json, '$.error') IS NOT NULL"

# Group results by cellsMatchResult
sqlite3 --header --column ${RESULTSDB} "SELECT json_extract(proto_json, '$.cellsMatchResult') as match_result, COUNT(*) as count FROM results GROUP BY match_result"
```

You can use the following query to look at the errors

```
sqlite3 ${RESULTSDB} "SELECT 
id, 
json_extract(proto_json, '$.error') as error 
FROM results WHERE json_extract(proto_json, '$.error') IS NOT NULL;"
```
