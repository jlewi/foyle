# Create a PR

* This is a recipe for creating PR descriptions using AI.
* Its a work in progress

1. Find the mergepoint with the main branch

```sh {"id":"01J4J3KN50MB5Z444F0BJAKQYR"}
git fetch origin
export FORKPOINT=$(git merge-base --fork-point origin/main)
```

2. Create a file with the log messages of all the changes up to this point

```sh {"id":"01J4J3PJJJEKCV17HVPTT40C3R"}
git log ${FORKPOINT}..HEAD > /tmp/commitlog
```

3. Use the llmtool to summarize the commit messages

```sh {"id":"01J4J3RFE8GYFK8EZZZ7ZV1MZ2"}
cat /tmp/commitlog  | llm "Here is the commit log for a bunch of messages. Please turn them into a detailed message suitable for the PR description for a PR containing all the changes. Do not include low value messages like 'fix lint' or 'fix tests'. Avoid superlatives and flowery language; just state what the change does and the reasoning behind it." > /tmp/commitsummary
cat /tmp/commitsummary
```

```sh {"id":"01J4J3WZN61KKNQ514AJAHFBQE","interactive":"false"}
cat /tmp/commitsummary
```

```sh {"id":"01J4J40733AKEC1A9ERV9CWZ5S","interactive":"false"}
REPOROOT=$(git rev-parse --show-toplevel)
cd ${REPOROOT}
CHANGEDFILES=$(git diff --name-only origin/main | grep -v -E '^protos/.*\.go$')
rm -f /tmp/changes.txt
while IFS= read -r file; do    
    git diff origin/main ${file} | llm "Here is the diff for a file. Create a bulleted list summarizing the changes. This list should be suitable for a git commit message or PR description" >> /tmp/changes.txt
    #echo "" >> /tmp/changes.txt
done <<< "${CHANGEDFILES}"

```

```sh {"id":"01J4J4R0FS9B93ADDB54HN4CNA","interactive":"false"}
cat /tmp/changes.txt
```