// @generated by protoc-gen-es v1.8.0 with parameter "target=ts,import_extension=none"
// @generated from file foyle/v1alpha1/agent.proto (syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";
import { Block, BlockOutput, Doc } from "./doc_pb";

/**
 * @generated from message GenerateRequest
 */
export class GenerateRequest extends Message<GenerateRequest> {
  /**
   * @generated from field: Doc doc = 1;
   */
  doc?: Doc;

  constructor(data?: PartialMessage<GenerateRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "GenerateRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "doc", kind: "message", T: Doc },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateRequest {
    return new GenerateRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateRequest {
    return new GenerateRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateRequest {
    return new GenerateRequest().fromJsonString(jsonString, options);
  }

  static equals(a: GenerateRequest | PlainMessage<GenerateRequest> | undefined, b: GenerateRequest | PlainMessage<GenerateRequest> | undefined): boolean {
    return proto3.util.equals(GenerateRequest, a, b);
  }
}

/**
 * @generated from message GenerateResponse
 */
export class GenerateResponse extends Message<GenerateResponse> {
  /**
   * @generated from field: repeated Block blocks = 1;
   */
  blocks: Block[] = [];

  constructor(data?: PartialMessage<GenerateResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "GenerateResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "blocks", kind: "message", T: Block, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateResponse {
    return new GenerateResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateResponse {
    return new GenerateResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateResponse {
    return new GenerateResponse().fromJsonString(jsonString, options);
  }

  static equals(a: GenerateResponse | PlainMessage<GenerateResponse> | undefined, b: GenerateResponse | PlainMessage<GenerateResponse> | undefined): boolean {
    return proto3.util.equals(GenerateResponse, a, b);
  }
}

/**
 * @generated from message ExecuteRequest
 */
export class ExecuteRequest extends Message<ExecuteRequest> {
  /**
   * @generated from field: Block block = 1;
   */
  block?: Block;

  constructor(data?: PartialMessage<ExecuteRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "ExecuteRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "block", kind: "message", T: Block },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ExecuteRequest {
    return new ExecuteRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ExecuteRequest {
    return new ExecuteRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ExecuteRequest {
    return new ExecuteRequest().fromJsonString(jsonString, options);
  }

  static equals(a: ExecuteRequest | PlainMessage<ExecuteRequest> | undefined, b: ExecuteRequest | PlainMessage<ExecuteRequest> | undefined): boolean {
    return proto3.util.equals(ExecuteRequest, a, b);
  }
}

/**
 * @generated from message ExecuteResponse
 */
export class ExecuteResponse extends Message<ExecuteResponse> {
  /**
   * @generated from field: repeated BlockOutput outputs = 1;
   */
  outputs: BlockOutput[] = [];

  constructor(data?: PartialMessage<ExecuteResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "ExecuteResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "outputs", kind: "message", T: BlockOutput, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ExecuteResponse {
    return new ExecuteResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ExecuteResponse {
    return new ExecuteResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ExecuteResponse {
    return new ExecuteResponse().fromJsonString(jsonString, options);
  }

  static equals(a: ExecuteResponse | PlainMessage<ExecuteResponse> | undefined, b: ExecuteResponse | PlainMessage<ExecuteResponse> | undefined): boolean {
    return proto3.util.equals(ExecuteResponse, a, b);
  }
}

