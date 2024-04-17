import * as docpb from "../gen/foyle/v1alpha1/doc_pb";

// CellMetadata is a dictionary of metadata that can be attached to a cell
export type CellMetadata = { [key: string]: any };

// getCellMetadata extracts the metadata from a block
export function getCellMetadata(block: docpb.Block): CellMetadata {  
  console.log(`getCellMetadata id=${block.id}`);
  return {
    "traceIds": block.traceIds,
    "id": block.id,
  };
}

// setBlockFromMeta is the inverse of getCellMetadata it populates the
// block with values from the metadata.
// TODO(jeremy): We need to do this for output blocks ass well
export function setBlockFromMeta(block: docpb.Block, meta: CellMetadata) {  
  // TODO(https://github.com/jlewi/foyle/issues/56). This was an attempt
  // to fix the error.
  if (meta === null) {
    console.log("meta is null not setting block");
    return;
  }
  if (meta === undefined) {
    console.log("meta is undefined");
    return;
  }
  if (Object.keys(meta).length === 0) {
    console.log("meta is an empty object");
  }
  if (meta.constructor !== Object) {
    console.log("Meta is not a plain object");
  }
  console.log("Meta is not undefined and is a plain object");
  if (Array.isArray(meta)) {
    console.log("meta is array");
    return;
  }
  console.log("Type of meta: ", typeof meta);  
  console.log("Value of meta: ", JSON.stringify(meta, null, 2));
  console.log("setBlockFromMetdata meta not null");
  if (meta.hasOwnProperty("traceIds")) {
    block.traceIds = meta["traceIds"];
    console.log("setBlockFromMetdata setting traceId");
  }
  if (meta.hasOwnProperty("id")) {
    console.log("setBlockFromMetdata setting block id");
    block.id = meta["id"];
  }
}