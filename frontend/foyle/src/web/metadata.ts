import * as docpb from "../gen/foyle/v1alpha1/doc_pb";

// CellMetadata is a dictionary of metadata that can be attached to a cell
export type CellMetadata = { [key: string]: any };

// getCellMetadata extracts the metadata from a block
export function getCellMetadata(block: docpb.Block): CellMetadata {  
  return {
    "traceIds": block.traceIds,
  };
}

// setBlockFromMeta is the inverse of getCellMetadata it populates the
// block with values from the metadata.
// TODO(jeremy): We need to do this for output blocks ass well
export function setBlockFromMeta(block: docpb.Block, meta: CellMetadata) {  
  if (meta.hasOwnProperty("traceIds")) {
    block.traceIds = meta["traceIds"];
  }
}