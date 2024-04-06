# VSCode Notebook APIs

The VSCode notebook APIs are a bit confusing. This document is an attempt to clarify them.

 Conversion routines between the protocol buffer format and VSCode's representation of a notebook

 VSCode has two different APIs
 1. Data APIs (NotebookData, NotebookCellData)
    https:github.com/microsoft/vscode/blob/98332892fd2cb3c948ced33f542698e20c6279b9/src/vs/workbench/api/common/extHostTypes.ts#L3598
    * These are concrete classes
    * This is what the serialization uses 
      * i.e. Serialize.serializeNotebook passes a VSCodeNotebookData as input
      * The deserializer returns a VSCodeNotebookData
    * NotebookCellData has an interface and concrete class
      concrete class - https:github.com/microsoft/vscode/blob/e4595ad3998a436bbf68d32a6c59e9d49fc63e32/src/vs/workbench/api/common/extHostTypes.ts#L3668
      interface - https:github.com/microsoft/vscode/blob/e4595ad3998a436bbf68d32a6c59e9d49fc63e32/src/vscode-dts/vscode.proposed.notebookMime.d.ts#L10

 2. Editor APIs (NotebookDocument, NotebookCell)
    * These are interfaces
    * These are used by the editor
    * To my knowledge there aren't concrete classes
 3. Both APIs use NotebookCellOutput, NotebookCellOutputItem
    * These are concrete classes
 4. NotebookCellData isn't compatible with the interface NotebookCell
    * NotebookCellData has extra fields such as index and document.

 However (NotebookCell and NotebookCellData) share the following fields:
    kind: NotebookCellKind
    metadata: [key: string];
    outputs: NotebookCellOutput[];

 However the differe in how the actual value is stored.

 NotebookCellData uses the fields:
    value: string;
    languageId: string;

 Whereas NotebookCell uses the field:
   document: TextDocument;

 TextDocument has the following fields and methods for our purposes
   languageId: string;
   getText(): string;

The TextDocument only appears to be used when reading cells from the editor.

The mutations API vscode.NotebookEdit.insertCells uses NotebookCellData.

Therefore in order to avoid repeating conversion code between VSCode Notebook classes and the protos we use the following patterns
* We have conversion methods to/from the Block proto to NotebookCellData
* We have conversion method from NotebookCell to NotebookCellData

