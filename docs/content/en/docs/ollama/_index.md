---
title: "Ollama"
description: "How to use Ollama with Foyle"
weight: 3
---

## What You'll Learn

How to configure Foyle to use models served by Ollama

## Prerequisites

1. Follow [Ollama's docs] to download Ollama and serve a model like `llama2`  

## Setup Foyle to use Ollama

Foyle relies on [Ollama's OpenAI Chat Compatability API]() to interact with models served by Ollama.


1. Configure Foyle to use the appropriate Ollama baseURL

   ```
   foyle config set openai.baseURL=http://localhost:11434/v1
   ```

   * Change the server and port to match how you are serving Ollama
   * You may also need to change the scheme to https; e.g. if you are using a VPN like [Tailscale](https://tailscale.com/)
    
1. Configure Foyle to use the appropriate Ollama model

   ```
   foyle config agent.model=llama2 
   ```
   
    * Change the model to match the model you are serving with Ollama

1. You can leave the `apiKeyFile` unset since you aren't using an API key with Ollama

1. That's it! You should now be able to use Foyle with Ollama
