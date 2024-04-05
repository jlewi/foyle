import * as vscode from "vscode";
import * as constants from './constants';
import * as agentpb from "../gen/foyle/v1alpha1/agent_pb";
import * as docpb from "../gen/foyle/v1alpha1/doc_pb";
import { extName } from "./constants";
// FoyleClient is a client for communicating with the backend.
//
// TODO(jeremy): One of 
export class FoyleClient {

    // TODO(jeremy): How do we deal with the async nature of the fetch call?
    // We should probably return a promise.
    public Execute(request: agentpb.ExecuteRequest): Promise<agentpb.ExecuteResponse> {
         // getConfiguration takes a section name.
         const config = vscode.workspace.getConfiguration(extName);
         // Include a default so that address is always well defined
         const address = config.get<string>("address", "http://localhost:8080"); 
        const resp = new agentpb.ExecuteResponse();
        const o = new docpb.BlockOutput();
        const i = new docpb.BlockOutputItem();
        i.textData = "some output";
        o.items = [i];
        resp.outputs = [
            o,  
        ];

        return fetch(address + "/api/v1alpha1/execute", {
            method: 'POST',
            body: resp.toJsonString(),
            headers: { 'Content-Type': 'application/json' },
        })
        .then(response => response.json())
        .then(data => {
            const resp = agentpb.ExecuteResponse.fromJson(data);
            return resp;
        });                
    }    
}

