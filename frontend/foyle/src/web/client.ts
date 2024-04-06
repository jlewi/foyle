import * as vscode from "vscode";
import * as constants from './constants';
import * as agentpb from "../gen/foyle/v1alpha1/agent_pb";
import * as docpb from "../gen/foyle/v1alpha1/doc_pb";
import { extName } from "./constants";
// FoyleClient is a client for communicating with the backend.
//
// TODO(jeremy): One of 
export class FoyleClient {

    // Execute a request.
    public Execute(request: agentpb.ExecuteRequest): Promise<agentpb.ExecuteResponse> {
        // getConfiguration takes a section name.
        const config = vscode.workspace.getConfiguration(extName);
        // Include a default so that address is always well defined
        const address = config.get<string>("executor-address", "http://localhost:8080"); 

        return fetch(address + "/api/v1alpha1/execute", {
            method: 'POST',
            body: request.toJsonString(),
            headers: { 'Content-Type': 'application/json' },
        })
        .then(response => response.json())
        .then(data => {
            const resp = agentpb.ExecuteResponse.fromJson(data);
            return resp;
        });                
    } 

    // generate a completion.
    public generate(request: agentpb.GenerateRequest): Promise<agentpb.GenerateResponse> {
        // getConfiguration takes a section name.
        const config = vscode.workspace.getConfiguration(extName);
        // Include a default so that address is always well defined
        const address = config.get<string>("agent-address", "http://localhost:8080"); 
        
        return fetch(address + "/api/v1alpha1/generate", {
           method: 'POST',
           body: request.toJsonString(),
           headers: { 'Content-Type': 'application/json' },
       })
       .then(response => response.json())
       .then(data => {
           const resp = agentpb.GenerateResponse.fromJson(data);
           return resp;
       });                
   }    
}


// getTraceID takes a list of blocks and returns the most recent traceId 
// if there is one. 
// Returns the empty string if there is no traceId
export function getTraceID(blocks: docpb.Block[]): string {
    if (blocks.length <= 0) {
      return "";
    }    
    var lastBlock = blocks.at(-1);
    if (lastBlock === undefined) {
      return "";
    }
    if (lastBlock.traceIds.length === 0) {
      return "";
    }
    const v = lastBlock.traceIds.at(-1);
    if (v === undefined) {
      return "";
    }
    return v;
  }