import * as vscode from "vscode";
import { promisify } from 'util';
import * as constants from './constants';
import * as converters from './converters';
import * as client from './client';
import * as agentpb from "../gen/foyle/v1alpha1/agent_pb";
import * as docpb from "../gen/foyle/v1alpha1/doc_pb";

// Controller handles execution of cells.
export class Controller {
  readonly label: string = "Foyle Notebook";  
  readonly notebookType: string = "foyle-notebook";
  readonly id: string = "foyle-notebook-kernel";

  // The controller needs to register the languages it handles. Each cell in the
  // notebook has a langId field. That is used to match it to the supported kernel.
  // I think that is used to match the controller to the cell.
  readonly supportedLanguages = [constants.bashLang];

  private client: client.FoyleClient;
  private readonly _controller: vscode.NotebookController;
  private _executionOrder = 0;    
  constructor(client: client.FoyleClient) {
    this.client = client;    
    this._controller = vscode.notebooks.createNotebookController(
      this.id,
      this.notebookType,
      this.label
    );

    this._controller.supportedLanguages = this.supportedLanguages;
    this._controller.supportsExecutionOrder = true;
    this._controller.executeHandler = this._executeAll.bind(this);
  }

  public dispose() {
    this._controller.dispose();
  }

  private _executeAll(
    cells: vscode.NotebookCell[],
    _notebook: vscode.NotebookDocument,
    _controller: vscode.NotebookController
  ): void {
    for (const cell of cells) {
      this._doExecution(cell);
    }
  }

  private async _doExecution(cell: vscode.NotebookCell): Promise<void> {
    console.log("Executing cell");
    const execution = this._controller.createNotebookCellExecution(cell);    
    execution.executionOrder = ++this._executionOrder;
    // Keep track of elapsed time to execute cell.
    execution.start(Date.now()); 
    
    // TODO(jeremy): Should we use the then syntax on callAwait
    const result = await callExecute(cell, this.client);

    execution.replaceOutput(result);
    execution.end(true, Date.now());
  }
}

async function callExecute(cell: vscode.NotebookCell, client: client.FoyleClient): Promise<vscode.NotebookCellOutput[]> {
  if (cell.metadata.hasOwnProperty("id")) {
    console.log(`callExecute called on block id = ${cell.metadata["id"]}`);
  } else {
    console.log(`callExecute called on block without id`);
  }
  console.log(`callExecute called ${cell.document.getText()}`); 

  const request = new agentpb.ExecuteRequest();

  request.block = converters.cellDataToBlock(converters.cellToCellData(cell));
  
  return client.Execute(request).then((response) => {
    let output: vscode.NotebookCellOutput[] = [];

    // TODO(jeremy): 
    // 1. Create a promisy that calls the backend
    // 2. Await that promise
    
    // TODO(jeremy): We want to support tracing. To do that we want to get the traceId from the response
    // and add it to the metadata of the input cell.
    try {        
      for (let outBlock of response.outputs) {
        let items: vscode.NotebookCellOutputItem[] = [];

        for (let outItem of outBlock.items) {
          if (outItem.textData) {
            const encoder = new TextEncoder();
            const uint8Array = encoder.encode(outItem.textData);
  
            items.push(new vscode.NotebookCellOutputItem(uint8Array, 'text/plain'));
          }
        }
        
        // TODO(jeremy): We need to add support for metadata containing the traceid.
        output.push(new vscode.NotebookCellOutput(items));
      }
    } catch (error) {
      console.error(`Failed to issue execute; Error: ${error}`);    
    }
  
    return output;
  });      
}
