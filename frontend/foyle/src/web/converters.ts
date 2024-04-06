// Conversion routines between the protocol buffer format and VSCode's representation of a notebook
//
// VSCode has two different APIs
// 1. Data APIs (NotebookData, NotebookCellData)
//    https://github.com/microsoft/vscode/blob/98332892fd2cb3c948ced33f542698e20c6279b9/src/vs/workbench/api/common/extHostTypes.ts#L3598
//    * These are concrete classes
//    * This is what the serialization uses 
//      * i.e. Serialize.serializeNotebook passes a VSCodeNotebookData as input
//      * The deserializer returns a VSCodeNotebookData
//    * NotebookCellData has an interface and concrete class
//      concrete class - https://github.com/microsoft/vscode/blob/e4595ad3998a436bbf68d32a6c59e9d49fc63e32/src/vs/workbench/api/common/extHostTypes.ts#L3668
//      interface - https://github.com/microsoft/vscode/blob/e4595ad3998a436bbf68d32a6c59e9d49fc63e32/src/vscode-dts/vscode.proposed.notebookMime.d.ts#L10
//
// 2. Editor APIs (NotebookDocument, NotebookCell)
//    * These are interfaces
//    * These are used by the editor
//    * To my knowledge there aren't concrete classes
// 3. Both APIs use NotebookCellOutput, NotebookCellOutputItem
//    * These are concrete classes
// 4. NotebookCellData isn't compatible with the interface NotebookCell
//    * NotebookCellData has extra fields such as index and document.
//
// However (NotebookCell and NotebookCellData) share the following fields:
//    kind: NotebookCellKind
//    metadata: [key: string];
//    outputs: NotebookCellOutput[];
//
// However the differe in how the actual value is stored.

// NotebookCellData uses the fields:
//    value: string;
//    languageId: string;
//
// Whereas NotebookCell uses the field:
//   document: TextDocument;
//
// TextDocument has the following fields and methods for our purposes
//   languageId: string;
//   getText(): string;
//
//

import * as vscode from 'vscode';
import * as docpb from "../gen/foyle/v1alpha1/doc_pb";
import * as metadata from "./metadata";
import * as constants from "./constants";

// cellToCellData converts a NotebookCell to a NotebookCellData.
// NotebookCell is an interface used by the editor.
// NotebookCellData is a concrete class.
export function cellToCellData(cell: vscode.NotebookCell): vscode.NotebookCellData {
  let data = new vscode.NotebookCellData(cell.kind, cell.document.getText(), cell.document.languageId);

  data.metadata = cell.metadata;
  data.outputs = [];
  for (let o of cell.outputs) {
    data.outputs.push(o);
  }
  return data;
}

// cellDataToBlock converts an instance of NotebookCellData to the Block proto.
export function cellDataToBlock(cell: vscode.NotebookCellData): docpb.Block {
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

  return block;
}

// blockToCellData converts a Block proto to a NotebookCellData.
export function blockToCellData(block: docpb.Block): vscode.NotebookCellData {
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
  return newCell;
}
