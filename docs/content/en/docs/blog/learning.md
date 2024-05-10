---
title: "Copilots for Devops; because you need to operate your service not just write it"
linkTitle: "Copilots for Devops"
date: 2024-05-10
author: "[Jeremy Lewi](https://lewi.us/about)"
type: blog
descrption: Foyle is an open source assistant to help software developers deal with the pain of devops. One of Foyle's central premises is that creating a UX that implicitly captures human feedback is critical to building AIs that effectively assist us with operations. This post describes how Foyle logs that feedback.
---

Since co-pilot launched in 2021, AI accelerated software development has become the norm. More importantly, as [Simon Willison argued at last year’s AI Engineer Summit ](https://simonwillison.net/2023/Oct/17/open-questions/)with AI there has never been an easier time  to learn to code. This means the population of people writing code is set to explode. All of this begs the question, who is going to operate all this software? While writing code is getting easier, our infrastructure is becoming more complex and harder to understand. Perhaps we shouldn’t be surprised that as the cost of writing software decreases, we see an explosion in the number of tools and abstractions increasing complexity; the expanding [CNCF landscape](https://landscape.cncf.io/) is a great illustration of this.

The only way to keep up with AI assisted coding is with AI assisted operations. While we are being flooded with copilots and autopilots for writing software, I think there has been much less progress with assistants for operations. I think this is because 1) everyone’s infrastructure is different; an outcome of the combinatorial complexity of today’s options and 2) there is no single system with a complete and up to date picture of a company’s infrastructure. 

Consider a problem that has bedeviled me ever since I started working on [Kubeflow](https://www.kubeflow.org/); “I just want to deploy Jupyter on my company’s Cloud and access it securely”? To begin to ask A’s (ChatGPT, Claude, Bard) for help we need a whole bunch of knowledge no new hire would know a priori; e.g. What do we use for compute, ECS, GKE, GCE? What are we using for VPN, tailscale, IAP, Cognito? How do we attach credentials to Jupyter so we can access internal data stores? What should we do for storage; Persistent disk File store?

The fundamental problem is mapping a user’s intent, “deploy Jupyter”, to the specific set of operations to achieve that within our organization. The current solution is to build [platforms](https://learn.microsoft.com/en-us/platform-engineering/what-is-platform-engineering) that create higher level abstractions that hopefully more closely map to user intent while hiding implementation details.  Unfortunately, building platforms is expensive and time consuming. I have talked to organizations with 100s of engineers building an internal developer platform (IDP). 

Foyle is an OSS project that aims to simplify software operations with AI. Foyle uses notebooks to create a UX that encourages developers to express intent as well as actions. By logging this data, Foyle is able to build models that predict the operations needed to achieve a given intent. This is a problem which LLMs are unquestionably good at.


# Demo 

<!-- cc_load_policy turns on captions by default-->
<iframe width="560" height="315" src="https://www.youtube.com/embed/RsHL8g6HYuA?cc_load_policy=1&si=XD9LZj_Yc6XDt6zW" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>


Let’s consider one of the most basic operations; fetching the logs to understand why something isn’t working. Observability is critical but at least for me a constant headache. Each observability tool has their own hard to remember query language and queries depend on how applications were instrumented. As an example, [Hydros](https://github.com/jlewi/hydros) is a tool I built for CICD. To figure out whether hydros successfully built an image or hydrated some manifests I need to query its logs. 

A convenient and easy way for me to express my intent is with the query


```
fetch the hydros logs for the image vscode-ext
```


If we send this to an AI (e.g. ChatGPT) with no knowledge of our infrastructure we get an answer which is completely wrong.


```
hydros logs vscode-ext
```


This is a good guess but wrong because my logs are stored in Google Cloud Logging. The correct query executed using gcloud is the following.


```
gcloud logging read 'logName="projects/foyle-dev/logs/hydros" jsonPayload.image="vscode-ext"' --freshness=1d  --project=foyle-dev

```


Now the very first time I access hydros logs I’m going to have to help Foyle understand that I’m using Cloud Logging and how hydros structures its logs; i.e. that each log entry contains a field image with the name of the image being logged. However, since Foyle logs the intent and final action it is able to learn. The next time I need to access logs if I issue a query like


```
show the hydros logs for the image caribou
```


Foyle responds with the correct query


```
gcloud logging read 'logName="projects/foyle-dev/logs/hydros" jsonPayload.image="caribou"' --freshness=1d --project=foyle-dev
```


I have intentionally asked for an image that doesn’t exist because I wanted to test whether Foyle is able to learn the correct pattern as opposed to simply memorizing commands. In this case, a single training example - on top of all the knowledge about gcloud embedded in the LLM- to answer correctly. With a single example Foyle learns 1) what log is used by hydros and 2) how hydros uses structured logging to associate log messages with a particular image.

Foyle relies on a UX which prioritizes collecting implicit feedback to improve the AI. 

![implicit feedback interaction diagram](../implicit_feedback_interaction_diagram.svg)

In this interaction, a user asks an AI to translate their intent into one or more tools the user can invoke. The tools are rendered in executable, editable cells inside the notebook. This experience allows the user to iterate on the commands if necessary to arrive at the correct answer. Foyle logs these iterations (see this previous [blog post](https://foyle.io/docs/blog/logfeedback/) for a detailed discussion) so it can learn from them.

The learning mechanism is quite simple. As denoted above we have the original query, Q, the initial answer from the AI, A, and then the final command, A’, the user executed. This gives us a triplet (Q, A, A’). If A=A’ the AI got the answer right; otherwise the AI made a mistake the user had to fix. 

The AI can easily learn from its mistakes by storing the pairs (Q, A’). Given a new query Qn we can easily search for similar queries from the past where the AI made a mistake. Matching text based on semantic similarity is one of the problems LLMs excel at. Using LLMs we can compute the embedding of Q and Qn and measure the similarity to find similar queries from the past. Given a set of similar examples from the past {(Q1,A1’),(Q2,A2’),...,(QN,AN’)} we can use few shot prompting to get the LLM to learn from those past examples and answer the new query correctly. As demonstrated by the example in the previous section this works quite well.

This pattern of collecting implicit human feedback and learning from it is becoming increasingly common. [Dosu](https://blog.langchain.dev/dosu-langsmith-no-prompt-eng/) uses this pattern to build AIs that can automatically label issues.


# An IDE For DevOps

One of the biggest barriers to building copilots for devops is that when it comes to operating infrastructure we are constantly switching between different modalities 



* We use IDEs/Editors to edit our IAC configs
* We use terminals to invoke CLIs
* We use UIs for click ops and visualization
* We use tickets/docs to capture intent and analysis
* We use proprietary web apps to get help from AIs

This fragmented experience for operations is a barrier to collecting data that would let us train powerful assistants. Compare this to writing software where a developer can use a single IDE to write, build, and test their software. When these systems are well instrumented you can train really valuable software assistants like Google’s [DIDACT](https://research.google/blog/large-sequence-models-for-software-development-activities/) and [Replit Code Repair](https://blog.replit.com/code-repair).

I think this is an opportunity to create a better experience for devops even in the absence of AI. A great example of this is what the [Runme.dev](https://runme.dev/) project is doing. Below is a screenshot of a Runme.dev interactive widget for VMs rendered directly in the notebook.


![runme gce renderer](../runme_gce_renderer.png)


This illustrates a UX where users don’t need to choose between the convenience of ClickOps and being able to log intent and action. Another great example is [Datadog Notebooks](https://docs.datadoghq.com/notebooks/). When I was at Primer, I found using Datadog notebooks to troubleshoot and document issues was far superior to copying and pasting links and images into tickets or Google Docs.  


# Conclusion: Leading the AI Wave

If you’re a platform engineer like me you’ve probably spent previous waves of AI building tools to support AI builders; e.g. by exposing GPUs or deploying critical applications like Jupyter. Now we, platform engineers, are in a position to use AI to solve our own problems and better serve our customers. Despite all the excitement about AI, there’s a shortage of examples of AI positively transforming how we work. Let's make platform engineering a success story for AI.

# Acknowledgements

I really appreciate [Hamel Husain](https://hamel.dev/) reviewing and editing this post.
