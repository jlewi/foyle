# Learning eval

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

1. We can use Foyle's logs to build a dataset set consisting of ground truth (query, response) (will abbreviate this as (q, r) pairs.
   * A query is the natural language query sent to Foyle.
   * A response is the command(s) that the user actually executed
1. We can use a powerful LLM like GPT4 or Claude to generate alternative query's for the same response
   *  f_{GPT4}(q) -> q'
1. Using the synthetically generated queries we can produce two data sets
   * A training set consisting of (q, r) pairs - these will be used for few shot prompting 
   * A test set consisting of (q', Agent(q')) pairs
1. To quantify the performance of the Agent can use an error metric similar to edit distance to measure the similarity 
   between the generated command and the actual command.
1. We can compare the performance of the Agent before and after learning from human feedback.

## Learning: What Do We Want To Learn

In the context of DevOps there are different types of learning we can test for. The simplest things are facts.
These are discrete pieces of information that can't be derived from other information. For example:

Query: "Which cluster is used for development?
Response: "gcloud container clusters describe --region=us-west1 --project=foyle-dev dev"

The fact that your organization is using a GKE cluster in the us-west1 region in project foyle-dev for development 
(and not an EKS cluster in us-est1) is something a teammate needs to tell you. In the context of our Agent, what
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

### Non-manual techniques 

In the future we will explore other techniques for generating evaluation sets that don't rely on each example
being human crafted. Two potential techniques are

* Using LLMs to generate synthetic data
* Using Foyle's logs

Both of these methods introduce additional challenges. If we use LLMs to generate synthetic data, how do we ensure
the queries and responses are valid? If we use Foyle's logs, how do we ensure the evaluation set isn't also part
of Foyle's training set?

## Evaluating correctness

We can evaluate the correctness of the system by comparing the generated commands with the actual commands.
We can use an error metric similar to edit distance. Rather than comparing individual characters, we can use
the following algorithm. Divide a command into positional and named arguments. For this purpose the command
itself is the first positional argument. For the positional we compute the edit distance but looking at the entirety
of the positional argument. For the named arguments we match arguments by name and count the number of incorrect,
missing, and extra arguments.

```math
D(i, j) = \begin{cases} 
0 & \text{if } i = 0 \text{ and } j = 0 \\
i & \text{if } j = 0 \text{ and } i > 0 \\
j & \text{if } i = 0 \text{ and } j > 0 \\
min \begin{cases} 
D(i-1, j) + 1 \\
D(i, j-1) + 1 \\
D(i-1, j-1) + 1 & \text{if } a[i] \neq b[j] \\
D(i-1, j-1) & \text{if } a[i] = b[j]
\end{cases} & \text{if } i, j > 0
\end{cases}
```

## How do we ensure the evaluation set isn't in our training set

For memorization we can check the exact query isn't in the training set. For generalization we need to make sure
the command isn't in the training set.



## Add an Option to filter out eval