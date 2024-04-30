---
runme:
  id: 01HWP5TGX8T7WR6FDJNH0T5MAY
  version: v3
---

# Learning Evaluation

* **Author**: Jeremy Lewi
* **Last Updated**: 2024-04-27
* **Status**: Being Drafted

## Objective

Measure the efficacy of the learning system.

## TL;DR

The key hypothesis of Foyle is that by using implicit human feedback, we can create an AI that automatically learns
about your infrastructure. In [TN002_Learning](tn002_learning) and [PR#83](https://github.com/jlewi/foyle/pull/83)
we implemented a very simple learning mechanism based on query dependent few shot prompting. The next step is to
quantiatively evaluate the effectiveness of this system. We'd like to construct an evaluation data set that consists of
queries whose answers depend on private knowledge of your infrastructure. We'd like to compare the performance of
Foyle before and after learning from human feedback. We propose to achieve this as follows

1. Manually construct an evaluation dataset consisting of queries whose answers depend on private knowledge of your
   infrastructure $\{(q_0, r_0), .... (q_n, r_n)\}$.
1. For each query $q_i$ in the evaluation dataset, we will generate a command $r'_i$ using the Agent. 
1. Compute an evaluation score using a metric similar to the edit distance 

$$
S = \sum_{i=0}^n D(r_i, r'_i) 
$$

1. Compare the scores before and after learning from human feedback.

## Learning: What Do We Want To Learn

In the context of DevOps there are different types of learning we can test for. The simplest things are facts.
These are discrete pieces of information that can't be derived from other information. For example:

Query: "Which cluster is used for development?

Response: "gcloud container clusters describe --region=us-west1 --project=foyle-dev dev"

The fact that your organization is using a GKE cluster in the us-west1 region in project foyle-dev for development
(and not an EKS cluster in us-west1) is something a teammate needs to tell you. In the context of our Agent, what
we want the agent to learn is alternative ways to ask the same question. For example,
"Show me the cluster where dev workloads run" should return the same response.

As a team's infrastructure grows, they often develop conventions for doing things in a consistent way. A simple
example of this is naming conventions. For example, as the number of docker images grows, most organizations develop
conventions around image tagging; e.g. using `live` or `prod` to indicate the production version of an image.
In the context of Foyle, we'd like to evaluate the efficacy with which Foyle learns these conventions from examples.
For example, our training set might include the example

Query: "Show me the latest version of the production image for hydros"

Response: "gcloud artifacts docker images describe us-west1-docker.pkg.dev/foyle-public/images/hydros/hydros:prod"

Our evaluation set might include the example

Query: "Describe the production image for foyle"

Response: "gcloud artifacts docker images describe us-west1-docker.pkg.dev/foyle-public/images/foyle/foyle:prod"

Notably, unlike the previous example, in this case both the query and response are different. In order for the agent
to generalize to images not in the training set, it needs to learn the convention that we use to tag images.

### Multi-step

We'd eventually like to be able to support multi-step queries. For example, a user might ask
"Show me the logs for the most recent image build for foyle". This requires two steps

1. Listing the images to find the sha of the most recent image
2. Querying for the logs of the image build with that sha

We leave this for future work.

## Building an Evaluation Set

We will start by hand crafting our evaluation set.

Our evaluation set will consist of a set of `.foyle documents` checked into source control in the [data](../data)
directory. The evaluation data set is specific to an organization and its infrastructure. We will use the infrastructure
of Foyle's development team as the basis for our evaluation set.

To test Foyle's ability to learn conventions we will include examples that exercise conventions for non-existent
infrastructure (e.g. images, deployments, etc...). Since they don't exist, a user wouldn't actually query for them
so they shouldn't appear in our logs.

We can automatically classify each examples in the evaluation set as either a memorization or generalization example
based on whether the response matches one of the responses in the training set. We can use the distance metric proposed
below to measure the similarity between the generated command and the actual command.

## Evaluating correctness

We can evaluate the correctness of the Agent by comparing the generated commands with the actual commands.
We can use an error metric similar to edit distance. Rather than comparing individual characters, we will compare arguments as a whole. 
First we will divide a command into positional and named arguments. For this purpose the command
itself is the first positional argument. For the positional we compute the edit distance but looking at the entirety
of the positional argument. For the named arguments we match arguments by name and count the number of incorrect,
missing, and extra arguments. We can denote this as follows. Let $a$ and $b$ be the sequence of positionsal
arguments for two different commands $r_a$ and $r_b$ that we wish to compare.

$$
a = {a_0, ..., a_{m}} \\
b = {b_0, ..., b_{n}}
$$

Then we can define the edit distance between the positional arguments as

$$
distance = D_p(m,n)
$$

$$
D_p(i,0) = \sum_{k=0}^{i} w_{del}(a_k)
$$

$$
D_p(0,j) = \sum_{k=0}^{j} w_{ins}(b_k)
$$

$$
D_p(i, j) = \begin{cases}
D_p(i-1, j-1) & \text{if } a_i = b_j \\
min \begin{cases}
D_p(i-1, j) + w_{del}(a_i) \\
D_p(i, j-1) + w_{ins}(b_j) \\
D_p(i-1, j-1) + w_{sub}(a_i, b_j)  \\
\end{cases} & \text{if } a[i] \neq b[j]
\end{cases}
$$

Here $w_{del}$, $w_{ins}$, and $w_{sub}$ are weights for deletion, insertion, and substitution respectively. If $w_{del} = w_{ins}$ then the distance is symetric $D(r_a, r_b) = D(r_b, r_a)$.

We can treat named arguments as two dictionaries `c` and `d`. We can define the edit distance between the named arguments as follows

$$
K = \text{keys}(c) \cup \text{keys}(d)
$$

$$
D_n = \sum_{k \in K} f(k)
$$

$$
f(k) = \begin{cases}
w_{del}(c[k]) & \text{if } k \notin \text{keys}(d) \\
w_{ins}(d[k]) & \text{if } k \notin \text{keys}(c) \\
w_{sub}(c[k], d[k]) & \text{if } k \in \text{keys}(c),  k \in \text{keys}(d), c[k] \neq d[k]  \\
0 & \text{otherwise}  \\
\end{cases}
$$

This definition doesn't properly account for the fact that named arguments often have a long and short version. We ignore this for now and accept that if one command uses the long version and the other uses the short version, we will count these as errors. In the future, we could try building a dictionary of known commands and normalizing the arguments to a standard form.

This definition also doesn't account for the fact that many CLIs have subcommands and certain named arguments must appear before or after certain subcommands. The definition above
would treat a named argument as a match as long it appears anywhere in the command.

The total distance between two commands is then

$$
D = D_p + D_n
$$

## Guardrails to avoid data leakage

We want to ensure that the evaluation set doesn't contain any queries that are in the training set. There are two cases we need to consider

1. Memorization: Queries are different but command is the same
   * These are valid training, evaluation pairs because we expect the command to exactly match even though the queries are different
1. Contamination: Queries are the same and commands are the same
   
We can use the distance metric proposed in the previous section to determine if a command in our training set matches the command in the eval dataset.

We can also use this to automatically classifiy each evaluation example as a memorization or generalization example. If the distance from the command in the evaluation set to the closest command in the training set is less than some threshold, we can classify it as a memorization example. Otherwise we can classify it as a generalization example.

## Add an Option to filter out eval

When processing an evaluation dataset we should probably denote in the logs that the request was part of the evaluation set. This way during the learning process we can filter it out to avoid learning our evaluation dataset.

An easy way to do this would be to include a log field `eval` that is set to `true` for all evaluation examples. When constructing the logger in [Agent.Generate](https://github.com/jlewi/foyle/blob/4a5e3b573cbfad34b3060eafff33bceeeef21636/app/pkg/agent/agent.go#L73) we can set this field to `true` if the Agent is in `eval` mode. 

We still want the Agent to log the evaluation examples so we can use the logs and traces to evaluate the Agent. 

In the future we could potentially set the `eval` field in the request if we want to use the same server for both training and evaluation. For now, we'd probably run the Agent as a batch job and not start a server.