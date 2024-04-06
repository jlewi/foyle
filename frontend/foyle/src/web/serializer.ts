import * as vscode from "vscode";
import * as docpb from "../gen/foyle/v1alpha1/doc_pb";
import * as constants from "./constants";
import * as metadata from "./metadata";
import * as converters from "./converters";

export class Serializer implements vscode.NotebookSerializer {
  // extensionPath should be the http path on which the extension is served
  private readonly extensionPath: string;
  
  constructor(extensionPath: string) {
    this.extensionPath = extensionPath;    
  }
  
  async deserializeNotebook(
    content: Uint8Array,
    _token: vscode.CancellationToken
  ): Promise<vscode.NotebookData> {
    
    var data = new TextDecoder().decode(content);
    data = data.trim();

    if (data === "") {
      // Since the file is empty we just return an empty notebook.
      // We don't want to try to parse it as json because that could potentially throw an error
      return new vscode.NotebookData([]);
    }

    let doc = docpb.Doc.fromJsonString(data);
    let cells: vscode.NotebookCellData[] = [];

    for (let block of doc.blocks) {
      let newCell = converters.blockToCellData(block);
      cells.push(newCell);
    }

      return new vscode.NotebookData(cells);    
  }

  async serializeNotebook(
    data: vscode.NotebookData,
    _token: vscode.CancellationToken
  ): Promise<Uint8Array> {
    console.log("serializing foyle doc");

    let doc = new docpb.Doc();
    doc.blocks = [];
    
    for (const cell of data.cells) {
      let block = converters.cellDataToBlock(cell);
      doc.blocks.push(block);
    }        
    
    const json = doc.toJsonString({prettySpaces: 2});
    console.log("serialized string length is: " + json.length);
    return new TextEncoder().encode(json);

  }
}