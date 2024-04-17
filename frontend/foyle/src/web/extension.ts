// The module 'vscode' contains the VS Code extensibility API
// Import the module and reference it with the alias vscode in your code below
import * as vscode from 'vscode';
import { Controller } from './controller';
import {FoyleClient} from './client';
import { Serializer } from './serializer';
import * as generate from './generate';
import * as debug from './debug';
// Create a client for the backend.
const client = new FoyleClient;

// This method is called when your extension is activated
// Your extension is activated the very first time the command is executed
export function activate(context: vscode.ExtensionContext) {
	console.log("Activating foyle extension uri:" + context.extensionUri);  
	context.subscriptions.push(
    vscode.workspace.registerNotebookSerializer(
      "foyle-notebook",
      new Serializer(context.extensionUri.path)
    )
  );
  
	// Register the controller for the notebook
  context.subscriptions.push(new Controller(client));
  context.subscriptions.push(new Controller(client, true));
  

  context.subscriptions.push(vscode.commands.registerCommand("foyle-notebook.newInteractive", async () => {
		const result: { inputUri: vscode.Uri, notebookUri?: vscode.Uri, notebookEditor?: vscode.NotebookEditor } | undefined = await vscode.commands.executeCommand('interactive.open',
			undefined,
			undefined,
			`${context.extension.id}/foyle-notebook-interactive-kernel`,
			undefined
		);
	}));

  // TODO(jeremy): Register a command to handle generation
  //context.subscriptions.push(vscode.commands.registerCommand("foyle.generate", handleGenerate));

	// Use the console to output diagnostic information (console.log) and errors (console.error)
	// This line of code will only be executed once when your extension is activated
	console.log('Congratulations foyle is now active in the web extension host!');

	// The command has been defined in the package.json file
	// Now provide the implementation of the command with registerCommand
	// The commandId parameter must match the command field in package.json
	let disposable = vscode.commands.registerCommand('foyle.helloWorld', () => {
		// The code you place here will be executed every time your command is executed
		// Display a message box to the user
		vscode.window.showInformationMessage('Hello World from foyle in a web extension host!');
	});

	// Here's where we register the command that will generate a completion using the AI model
	// You can set a keybinding for this command in the package.json file
  context.subscriptions.push(vscode.commands.registerCommand("foyle.generate", generate.generateCompletion));

	context.subscriptions.push(vscode.commands.registerCommand("foyle.printCell", debug.printCell));
	context.subscriptions.push(disposable);
}

// This method is called when your extension is deactivated
export function deactivate() {}
