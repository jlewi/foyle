import * as vscode from 'vscode';
import {FoyleClient, getTraceID} from './client';
import * as converters from './converters';
import * as docpb from "../gen/foyle/v1alpha1/doc_pb";
import * as agentpb from "../gen/foyle/v1alpha1/agent_pb";
// printCell is a debug tool that prints the contents of a cell to the console
// Its purpose is to help us debug problems by printing the contents of a cell
// In particular the cell metadata
export async function printCell() {
  const editor = vscode.window.activeNotebookEditor;

  if (!editor) {    
    return;
  }

  if (editor?.selection.isEmpty) {    
    return;
  }
  
  // We subtract 1 because end is non-inclusive
  const lastSelectedCell = editor?.selection.end - 1;
  var lastActiveCell = editor?.notebook.cellAt(lastSelectedCell);  
  console.log(`Cell value: ${lastActiveCell.document.getText()}`);
  console.log(`Cell language: ${lastActiveCell.document.languageId}`);
  console.log(`Cell kind: ${lastActiveCell.kind}`);
  
  Object.keys(lastActiveCell.metadata).forEach(key => {
    console.log(`Cell Metadata key: ${key} value: ${lastActiveCell.metadata[key]}`);
  });
}