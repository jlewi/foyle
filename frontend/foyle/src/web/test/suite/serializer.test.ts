import * as assert from 'assert';
import * as vscode from 'vscode';
import * as serializer from '../../serializer';
import * as docpb from "../../../gen/foyle/v1alpha1/doc_pb";
import { bashLang } from '../../constants';

class SerializationTestCase {
    public jsonDoc: any;
    public nbData: vscode.NotebookData;

    constructor(jsonDoc: any, nbData: vscode.NotebookData) {
        this.jsonDoc = jsonDoc;
        this.nbData = nbData;
    }
}

// construct a test case of the JSON and corresponding notebook cell data
function createTestCase(): SerializationTestCase {
    // TODO(jeremy): The test case is missing coverage for output blocks.
    // Doc to deserialize
    const jsonDoc = {
        "blocks": [
            {
              "kind": "MARKUP",
              "contents": "a markdown cell",
              "traceIds": ["a", "b"],
          },
            {
                "kind": "CODE",
                "language": "bash",
                "contents": "echo hello world",
            }              
        ]
    };    
    
    let cell1 =new vscode.NotebookCellData(vscode.NotebookCellKind.Markup, "a markdown cell", "");
    cell1.metadata = {
        "traceIds": ["a", "b"],
        "id": "",
    };

    let cell2 = new vscode.NotebookCellData(vscode.NotebookCellKind.Code, "echo hello world", bashLang);
    cell2.metadata = {
        "traceIds": [],
        "id": "",
    };

    const data = new vscode.NotebookData([
        cell1,
        cell2,
    ]);

    return new SerializationTestCase(jsonDoc, data);
}

suite('Serialization Test Suite', () => {
  test('deserialize', async () => {      
      const testCase = createTestCase();
      // Encode the json
      const jsonStr = JSON.stringify(testCase.jsonDoc);
      const jsonBytes = new TextEncoder().encode(jsonStr);

      const s = new serializer.Serializer("");
      const actual = await s.deserializeNotebook(jsonBytes, new vscode.CancellationTokenSource().token);
         
      assert.deepEqual(actual, testCase.nbData);
  });

  test('test empty-doc deserialize', async () => {
      // This test case ensures that we can deserialize an empty document.
      const jsonBytes: Uint8Array = new TextEncoder().encode("");

      const s = new serializer.Serializer("");
      const actual = await s.deserializeNotebook(jsonBytes, new vscode.CancellationTokenSource().token);

      const expected = new vscode.NotebookData([]);

      assert.deepEqual(actual, expected);
  });

  test('test serialize', async () => {
    const testCase = createTestCase();  

    const s = new serializer.Serializer("");

    const jsonBytes = await s.serializeNotebook(testCase.nbData, new vscode.CancellationTokenSource().token);
    const jsonStr = new TextDecoder().decode(jsonBytes);
    const actual = JSON.parse(jsonStr);

    assert.deepEqual(actual, testCase.jsonDoc);
    });
});