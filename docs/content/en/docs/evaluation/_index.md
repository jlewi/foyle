---
title: "Evaluation"
description: "How to evaluate Foyle's performance"
weight: 3
---

## What You'll Learn

How to quantitatively evaluate Foyle's AI quality

## Building an Evaluation Dataset

To evaluate Foyle's performance, you need to build an evaluation dataset. This dataset should contain a set of examples 
where you know the correct answer. Each example should be a `.foyle` file that ends with a code block containing a
command. The last command represents the command you'd like Foyle to predict when the input is the rest of the document.

Its usually a good idea to start by hand crafting a few examples based on knowledge of your infrastructure
and the kinds of questions you expect to ask. A good place to start is by creating examples that test basic knowledge
of how your infrastructure is configured e.g.

1. Describe the cluster used for dev/prod?
1. List the service accounts used for dev
1. List image repositories

A common theme here is to think about what information a new hire would need to know to get started. Even a new hire
experienced in the technologies you use would need to be informed about your specific configuration; e.g.
what AWS Account/GCP project to use.

For the next set of examples, you can start to think about conventions your teams use. For example, if you have 
specific tagging conventions for images or labels you should come up with examples that test whether Foyle has
learned those conventions

1. Describe the prod image for service foo?
1. Fetch the logs for the service bar?
1. List the deployment for application foo?

A good way to test whether Foyle has learned these conventions is to create examples for non-existent services, images,
etc... This makes it likely users haven't actually issued those requests causing Foyle to memorize the answers. 

## Evaluating Foyle

Once you have an evaluation dataset define an Experiment resource in a YAML file. Here's an example

```yaml
kind: Experiment
metadata:
  name: "learning"
spec:
  evalDir: /Users/jlewi/git_foyle/data/eval
  dbDir: /Users/jlewi/foyle_experiments/learning
  sheetID: "1O0thD-p9DBF4G_shGMniivBB3pdaYifgSzWXBxELKqE"
  sheetName: "Results"
  Agent:
    model: gpt-3.5-turbo-0125
    rag:
      enabled: true
      maxResults: 3
```

You can then run evaluation by running

```bash
foyle apply /path/to/experiment.yaml
```

This will run evaluation. If evaluation succeeds it will create a Google Sheet with the results. Each row in the
Google Sheet contains the results for one example in the evaluation dataset. The distance column measures
the distance between the predicted command and the actual command. The distance is the number of arguments
that would need to be changed to turn the predicted command into the actual command. If the actual and expected
are the same the distance will be zero. The maximum value of the edit distance is the maximum of the number of arguments
in the actual and expected commands. 
