import * as vscode from "vscode";
import * as docpb from "../gen/foyle/v1alpha1/doc_pb";
import * as constants from "./constants";
import * as metadata from "./metadata";
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
      let kind = vscode.NotebookCellKind.Markup;
      if (block.kind === docpb.BlockKind.CODE) { 
        kind = vscode.NotebookCellKind.Code;
      }
      let language = block.language;
      if (language === "" && kind === vscode.NotebookCellKind.Code) {
        // Default to the bash language
        language = constants.bashLang;

        if (language !== constants.bashLang) {
          console.log(`Unsupported language: ${language}`);
        }
      }

      let newCell = new vscode.NotebookCellData(
        kind,
        block.contents,
        language             
      );

      newCell.metadata = metadata.getCellMetadata(block);
      newCell.outputs = [];
      for (let output of block.outputs) {
        let items: vscode.NotebookCellOutputItem[] = [];
        for (let item of output.items) {
          if (item.textData !== "") {
            const text = new TextEncoder().encode(item.textData);            
            items.push(new vscode.NotebookCellOutputItem(text, item.mime));          
          } else {
            console.log("Unknown output type");
          }
        }
        newCell.outputs.push(new vscode.NotebookCellOutput(items));
      }
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
      let block = new docpb.Block();
      block.contents = cell.value;
      block.language = cell.languageId;
      if (cell.kind === vscode.NotebookCellKind.Code) {
        block.kind = docpb.BlockKind.CODE;
      } else {
        block.kind = docpb.BlockKind.MARKUP;
      }

      if (cell.metadata !== undefined) {
        metadata.setBlockFromMeta(block, cell.metadata);
      }
      
      let cellOutputs: vscode.NotebookCellOutput[] = [];
      if (cell.outputs !== undefined) {
        cellOutputs = cell.outputs;
      }
      for (const output of cellOutputs) {
        let outBlock = new docpb.BlockOutput();
        outBlock.items = [];
        for (const item of output.items) {
          let outItem = new docpb.BlockOutputItem();
          outItem.textData = new TextDecoder().decode(item.data);
          outItem.mime = item.mime;
          outBlock.items.push(outItem);
        }
        block.outputs.push(outBlock);
      }
      doc.blocks.push(block);
    }        

    //const json = JSON.stringify(doc.toJsonString, undefined, "  ");
    
    const json = doc.toJsonString({prettySpaces: 2});
    console.log("serialized string length is: " + json.length);
    return new TextEncoder().encode(json);

  }
}