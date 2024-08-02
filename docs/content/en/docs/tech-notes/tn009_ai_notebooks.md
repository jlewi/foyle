---
title: TN009 Review AI in Notebooks
description: Review AI in Notebooks
weight: 9
author: "[Jeremy Lewi](https://lewi.us/about)"
date: 2024-07-05
status: published
---

## Objective

* Review how other products/projects are creating AI enabled experiences in a notebook UX

## TL;DR

There are lots of notebook products and notebook like interfaces. This doc tries to summarize how they are using AI to 
enhance the notebook experience. The goal is to identify features and patterns that could
be useful incorporating into Foyle and RunMe.


## JupyterLab AI

[JupyterLab AI](https://github.com/jupyterlab/jupyter-ai?tab=readme-ov-file) is a project that aims to bring AI 
capabilities to JupyterLab. It looks like it introduces a couple ways of interacting with AI's

* GitHub copilot like chat window [docs](https://jupyter-ai.readthedocs.io/en/latest/users/index.html#the-chat-interface).
* [AI Magic Commands](https://jupyter-ai.readthedocs.io/en/latest/users/index.html#the-ai-and-ai-magic-commands)
* [Inline Completer](https://github.com/simonw/llm) uses ghost text to render completions to the current cell

The chat feature looks very similar to GitHub copilot chat. Presumably its more aware of the notebook structure.
For example, it lets you select notebook cells and include them in your prompt 
([docs](https://jupyter-ai.readthedocs.io/en/latest/users/index.html#asking-about-something-in-your-notebook)).

Chat offers built in RAG. Using the 
[/learn](https://jupyter-ai.readthedocs.io/en/latest/users/index.html#learning-about-local-data) command you can
inject a bunch of documents into a datastore. These documents are then used to answer questions in the chat using
RAG.

The [AI magic commands](https://jupyter-ai.readthedocs.io/en/latest/users/index.html#the-ai-and-ai-magic-commands) 
let you directly invoke different AI models. The prompt is the text inside the cell. The model output
is then rendered in a new cell. This is useful if your asking the model to produce code because you can then execute
the code in the notebook.

In RunMe you achieve much of the same functionality as the magics by using a CLI like [llm](https://github.com/simonw/llm)
and just executing that in a bash cell.

Compared to what's proposed in [TN008 Autocompleting cells](tn008_autocompleting_cells.md) it looks like the inline
completer only autocompletes the current cell; it doesn't suggest new cells. 

### JupyterLab AI Links

* [Early demo video of inline completion](https://github.com/jupyterlab/jupyterlab/issues/14267#issuecomment-1778365528)


## Hex

Hex magic lets you describe in natural language what you want to do and then it adds the cells to the notebook to do that.
For example, you can describe a graph you want and it will add an SQL cell to select the data
[Demo Video](https://hex.tech/blog/magic-ai/).

Notably, from a UX perspective prompting the AI is a "side conversation". There is a hover window that lets you 
ask the AI to do something and then the AI will modify the notebook accordingly. For example, in their
[Demo Video](https://hex.tech/blog/magic-ai/) you explicitly ask the AI to 
"Build a cohort analysis to look at customer churn". The AI then adds the cells to the notebook.
This is different from an AutoComplete like UX as proposed in [TN008 Autocompleting cells](tn008_autocompleting_cells.md).

In an Autocomplete like UX, a user might add a markdown cell containing a heading like "Customer Churn: Cohort Analysis".
The AI would then autocomplete these by inserting the cells to perform the analysis and render it. The user
wouldn't have to explicitly prompt the AI. 
