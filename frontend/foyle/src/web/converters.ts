// Conversion routines between the protocol buffer format and VSCode's representation of a notebook
//
// See ../vscode_apis.md for an exlanation. It is very helpful for understanding this folder.

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
