---
description: Getting started with Foyle Locally
title: Getting Started Locally
weight: 2
---

{{% pageinfo %}}
Deploying Foyle locally is the quickest, easiest way to try out Foyle.
{{% /pageinfo %}}

## Installation

### Prerequisites: VSCode & RunMe 

Foyle relies on [VSCode](https://code.visualstudio.com/) and [Runme.dev](https://runme.dev/)
to provide the frontend.

1. If you don't have VSCode visit the [downloads page](https://code.visualstudio.com/) and install it
1. Follow [Runme.dev](https://docs.runme.dev/installation/installrunme#installing-runme-on-vs-code) instructions to install the RunMe.dev extension in vscode

### Install Foyle

1. Download the latest release from the [releases page](https://github.com/jlewi/foyle/releases)

1. On Mac you may need to remove the quarantine attribute from the binary

   ```bash
   xattr -d com.apple.quarantine /path/to/foyle
   ```

## Setup

1. Configure your OpenAPI key

   ```sh
   foyle config set openai.apiKeyFile=/path/to/openai/apikey
   ```

   * If you don't have a key, go to [OpenAI](https://openai.com/) to
      obtain one

1. Start the server

   ```bash
   foyle serve
   ```

   * By default foyle uses port 8877 for the http server

   * If you need to use a different port you can configure this as follows

   ```sh
   foyle config set server.httpPort=<YOUR HTTP PORT>
   ```

1. Inside VSCode configure RunMe to use Foyle
   1. Open the VSCode setting palette
   1. Search for `Runme: Ai Base URL`
   1. Set the address to `http://localhost:${HTTP_PORT}/api`
      * The default port is 8877
      * If you set a non default value then it will be the value of `server.httpPort`

## Try it out!

Now that foyle is running you can open markdown documents in VSCode and start interacting with foyle.

1. Inside VSCode Open a markdown file or create a notebook; this will open the notebook inside RunMe
   * Refer to [RunMe's documentation](https://docs.runme.dev/installation/installrunme#full-display-of-runmes-action-on-a-markdown-file-in-vs-code) for a walk through
     of RunMe's UI
   * If the notebook doesn't open in RunMe
      1. right click on the file and select "Open With"
      1. Select the option "Run your markdown" to open it with RunMe
1. You can now add code and notebook cells like you normally would in vscode
1. To ask Foyle for help do one of the following

   * Open the command pallet and search for `Foyle generate a completion`
   * Use the shortcut key:
      * "win;" - on windows
      * "cmd;" - on mac

## Customizing Foyle VSCode Extension Settings

### Customizing the Foyle Server Address

1. Open the settings panel; you can click the gear icon in the lower left window and then select settings
2. Search for `Runme: Ai base URL
3. Set `Runme: Ai base URL` to the address of the Foyle server to use as the Agent
   * The Agent handles requests to generate completions

### Customizing the keybindings

If you are unhappy with the default key binding to generate a completion you can change it as follows

1. Click the gear icon in the lower left window and then select "Keyboard shortcuts"
2. Search for "Foyle"
3. Change the keybinding for `Foyle: generate a completion` to the keybinding you want to use

## Where to go next

* [Learning from feedback](/docs/learning/) - How to improve the AI with human feedback.