---
description: Search Gmail and classify and extract information
title: Email Processing
weight: 15
---

## What You'll Learn

This guide walks you through

* Searching Gmail using the [gctl CLI](https://github.com/jlewi/gctl)
* Downloading Gmail with [gctl CLI](https://github.com/jlewi/gctl)
* Extracting information from emails using the [llm CLI](https://github.com/simonw/llm)

For a demo video see [here](https://x.com/jeremylewi/status/1830662143374696738).

This document is intended to be opened and run in [RunMe](https://runme.dev/).

## Setup

Clone the [Foyle repository](https://github.com/jlewi/foyle) to get a copy of the document.

```sh
git clone https://github.com/jlewi/foyle /tmp/foyle
```

Open the file `foyle/docs/content/en/docs/use-cases/search_and_extract_gmail.md` in RunMe.

## Install Prerequisites

Install the [llm CLI](https://github.com/simonw/llm)

Using pip

```sh
pip install llm
```

Download the latest [release of gctl](https://github.com/jlewi/gctl/releases/latest)

```sh
TAG=$(curl -s https://api.github.com/repos/jlewi/gctl/releases/latest | jq -r '.tag_name')
# Remove the leading v because its not part of the binary name
TAGNOV=${TAG#v}
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
echo latest tag is $TAG
echo OS is $OS
echo Arch is $ARCH
LINK=https://github.com/jlewi/gctl/releases/download/${TAG}/gctl_${TAGNOV}_${OS}_${ARCH}
echo Downloading $LINK
wget $LINK -O /tmp/gctl
```

Move gctl onto your PATH

```bash
chmod a+rx /usr/local/bin/gctl
sudo mv /tmp/gctl /usr/local/bin/gctl
```

## Search For Emails

Search for emails using the gctl CLI

* Change the search query to match your use case
* You can refer to the [Gmail search documentation](https://developers.google.com/gmail/api/guides/filtering)
* Or if you've enabled Foyle you can just describe the search query in natural language in a markdown cell and let Foyle figure out
   the syntax for you!

```bash
gctl mail search "kamala" > /tmp/list.json
cat /tmp/list.json
```

## Extract Information From The Emails

* Describe a simple program which uses the gctl CLI to read the email messages and then extracts information using the LLM tool
* Then let Foyle generate the code for you!

Here is an example program that generated the code below:

* Loop over the json dictionaries in /tmp/list.json
* Each dictionary has an ID field
* Use the ID field to read the email using the gctl command
* Save the email to the file /tmp/message_${ID}.json
* Pipe the email to the llm tool
* Use the llm tool to determine if the email is from the Kamala Harris campaign. If it is extract the unsubscribe link. Emit the data as a JSON dictionary with the fields
   * from
   * isKamala
   * usubscribeLink

```bash
#!/bin/bash
for id in $(jq -r '.[].ID' /tmp/list.json); do
  gctl mail get "$id" > "/tmp/message_${id}.json"
  cat "/tmp/message_${id}.json" | llm "Is this email from the Kamala Harris campaign? If yes, extract the unsubscribe link. Output a JSON with fields: from, isKamala (true/false), unsubscribeLink (or null if not found)" 
done
```