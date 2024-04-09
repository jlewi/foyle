---
title: "Azure OpenAI"
description: "Using Azure OpenAI with Foyle"
weight: 3
---

## What You'll Learn

How to configure Foyle to use Azure OpenAI

## Prerequisites

1. You need an Azure Account (Subscription) 
1. You need access to [Azure Open AI](https://learn.microsoft.com/en-us/azure/ai-services/openai/overview#how-do-i-get-access-to-azure-openai)


## Setup Azure OpenAI

You need the following Azure OpenAI resources:

*  [Azure Resource Group](https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/manage-resource-groups-portal) - This will be an Azure resource group that contains your Azure OpenAI resources    

*  [Azure OpenAI Resource Group](https://learn.microsoft.com/en-us/azure/ai-services/openai/how-to/create-resource?pivots=web-portal) - This will contain your Azure OpenAI model deployments 
   
   * You can use the Azure CLI to check if you have the required resources
   
     ```
     az cognitiveservices account list --output=table
     Kind    Location    Name            ResourceGroup
     ------  ----------  --------------  ----------------
     OpenAI  eastus      ResourceName    ResourceGroup
     ```

   * **Note** You can use the [pricing page](https://azure.microsoft.com/en-us/pricing/details/cognitive-services/openai-service/) to see which models are available
     in a given region. Not all models are available an all regions so you need to select a region with the models you want to use with Foyle. 
   * Foyle currently uses [gpt-3.5-turbo-0125](https://platform.openai.com/docs/models/gpt-3-5-turbo)

* A GPT3.5 deployment 
  * Use the CLI to list your current deployments
    
    ```
    az cognitiveservices account deployment list -g ${RESOURCEGROUP} -n ${RESOURCENAME} --output=table
    ```

  * If you need to create a deployment follow the [instructions](https://learn.microsoft.com/en-us/azure/ai-services/openai/how-to/create-resource?pivots=web-portal#deploy-a-model) 

## Setup Foyle To Use Azure Open AI

### Set the Azure Open AI BaseURL

We need to configure Foyle to use the appropriate Azure OpenAI endpoint. You can use the [CLI](https://learn.microsoft.com/en-us/azure/ai-services/openai/how-to/create-resource?pivots=cli#get-the-endpoint-url) to determine
the endpoint associated with your resource group 

```
az cognitiveservices account show \
--name <myResourceName> \
--resource-group  <myResourceGroupName> \
| jq -r .properties.endpoint
```

Update the baseURL in your Foyle configuration

```
foyle config set azureOpenAI.baseURL=https://endpoint-for-Azure-OpenAI
```

### Set the Azure Open AI API Key

Use the CLI to obtain the API key for your Azure OpenAI resource and save it to a file

```
az cognitiveservices account keys list \
--name <myResourceName> \
--resource-group  <myResourceGroupName> \
| jq -r .key1  > ${HOME}/secrets/azureopenai.key
```

Next, configure Foyle to use this API key

```
foyle config set azureOpenAI.apiKeyFile=/path/to/your/key/file
```

### Specify model deployments 

You need to configure Foyle to use the appropriate Azure deployments for the models Foyle uses.

Start by using the Azure CLI to list your deployments

  ```
  az cognitiveservices account deployment list --name=${RESOURCE_NAME} --resource-group=${RESOURCE_GROUP} --output=table
  ```

Configure Foyle to use the appropriate deployments

```
foyle config set azureOpenAI.deployments=gpt-3.5-turbo-0125=<YOUR-GPT3.5-DEPLOYMENT-NAME>
```

### Troubleshooting: 

#### Rate Limits

If Foyle is returning rate limiting errors from Azure OpenAI, use the CLI to check 
the rate limits for your deployments

```
az cognitiveservices account deployment list -g ${RESOURCEGROUP} -n ${RESOURCENAME}
```

Azure OpenAI sets the default values to be quite low; 1K tokens per minute. This is usually much
lower than your allotted quota. If you have available quota, you can use the UI or to increase these limits.
