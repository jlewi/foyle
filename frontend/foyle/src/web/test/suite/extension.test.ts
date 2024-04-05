import * as assert from 'assert';

// You can import and use all API from the 'vscode' module
// as well as import your extension to test it
import * as vscode from 'vscode';
import * as client from '../../client';
import * as agentpb from '../../../gen/foyle/v1alpha1/agent_pb';

suite('Web Extension Test Suite', () => {
	vscode.window.showInformationMessage('Start all tests.');

	test ('client test', async () => {
		// N.B. This is not really a unittest. It is testing we can send a request.
		// TODO(jeremy): This could probably be improved.
		const c = new client.FoyleClient();
		try {
		const resp = await c.Execute(new agentpb.ExecuteRequest());
			console.log("Got response");
			console.log(resp);
			vscode.window.showInformationMessage('success.');		
		}	catch(err){
			vscode.window.showInformationMessage('failure.');
			console.log("Got error");
		};		
	});
	test('Sample test', () => {
		assert.strictEqual(-1, [1, 2, 3].indexOf(5));
		assert.strictEqual(-1, [1, 2, 3].indexOf(0));
	});
});
