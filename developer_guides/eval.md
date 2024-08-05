## Running Evaluation

## Running Level 1 Evaluation

Level 1 evaluations are assertions that run on AI responses.

To evaluate changes to the agent first setup an instance of the agent with the changes you want.
Be sure to configure it so that it stores logs and responses in a different directory than your production
agent because you don't want the evaluation data to contaminate the learning process.

```sh {"id":"01J4DJT0G24YH9K4F8YRTSZD8N"}
export REPOROOT=$(git rev-parse --show-toplevel)
export RUNDIR=${REPOROOT}/experiments/runs/$(date +%Y%m%d_%H%M%S)
echo "Using run directory: ${RUNDIR}"
```

### Setup the configuration for the agent in this run

```sh {"id":"01J4DKE3M85ETKNHFH4G0HT0M6"}
mkdir -p ${RUNDIR}
cp ~/.foyle/config.yaml ${RUNDIR}/config.yaml
```

* Adjust the ports used by the agent to avoid conflicts with the production agent

```sh {"id":"01J4DKK0N36XN2HV4GQK7YRXCC"}
yq e '.server.httpPort = 55080' -i ${RUNDIR}/config.yaml
yq e '.server.grpcPort = 55090' -i ${RUNDIR}/config.yaml
```

* We make a copy of the training directory to a new directory for this evaluation run.

```sh {"id":"01J4DKP9P59GCGNG6QXX6KR9AF"}
cp -r ~/.foyle/training ${RUNDIR}/
```

```sh {"id":"01J4DKQXXB8P7CV7VS4YS5DHDD"}
yq e ".learner.exampleDirs=[\"${RUNDIR}/training\"]" -i ${RUNDIR}/config.yaml
```

* Remove the RunMe directory for the extra log directory
* We don't want to reprocess RunMe logs
* Since we aren't actually using the Frontend there are no RunMe logs to process anyway

```sh {"id":"01J4F79ZE8YAAKV252G2T7XD25"}
yq e ".learner.logDirs=[]" -i ${RUNDIR}/config.yaml
```

* configure the assertions

```sh {"id":"01J4F896JP8FZ3N8BGVPZ7VHJ4"}
cp -f ${REPOROOT}/experiments/assertions.yaml ${RUNDIR}/assertions.yaml
yq e ".spec.agentAddress=http://localhost:55080/api" -i ${RUNDIR}/assertions.yaml
yq e ".spec.dbDir=\"${RUNDIR}/evalDB\"" -i ${RUNDIR}/assertions.yaml

```

### Run the agent

* Start the agent containing the changes you want to evaluate

```sh {"id":"01J4DM107F0GJWJKFV4P77TAQY"}
cd ${REPOROOT}/app
export CONFIGFILE=${RUNDIR}/config.yaml
go run github.com/jlewi/foyle/app serve --config=${CONFIGFILE}
```

### Run evaluation driver

```sh {"id":"01J4F8KQ7N5DE3JQRX33T60BB0"}
cd ${REPOROOT}/app
export CONFIGFILE=${RUNDIR}/config.yaml
go run github.com/jlewi/foyle/app apply --config=${CONFIGFILE} ${RUNDIR}/assertions.yaml
```

### Analyze the results

```sh {"id":"01J4HJY2M13P3X60N9WG9BCSTV"}
ls -la ${RUNDIR}
```

```sh {"id":"01J4HN72G5EY98MYPCZG7V02WZ","interactive":"false","mimeType":"application/json"}
curl -s -H "Content-Type: application/json" http://localhost:55080/api/EvalService/AssertionTable -d "{\"database\":\"${RUNDIR}/evalDB\"}" | jq .rows
```

## Run baseline experiment

```sh {"id":"01HZ38BC6WJF5RB9ZYTXBJE38M"}
foyle apply ~/git_foyle/experiments/norag.yaml
```

## Run experiment with RAG

```sh {"id":"01HZ38QWPZ565XH11CCKYCF1M7"}
foyle apply ~/git_foyle/experiments/rag.yaml
```