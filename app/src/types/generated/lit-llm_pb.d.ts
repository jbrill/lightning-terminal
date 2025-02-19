// package: litrpc
// file: lit-llm.proto

import * as jspb from "google-protobuf";

export class AnalyzeNodeRequest extends jspb.Message {
  getQuery(): string;
  setQuery(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AnalyzeNodeRequest.AsObject;
  static toObject(includeInstance: boolean, msg: AnalyzeNodeRequest): AnalyzeNodeRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: AnalyzeNodeRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AnalyzeNodeRequest;
  static deserializeBinaryFromReader(message: AnalyzeNodeRequest, reader: jspb.BinaryReader): AnalyzeNodeRequest;
}

export namespace AnalyzeNodeRequest {
  export type AsObject = {
    query: string,
  }
}

export class AnalyzeNodeResponse extends jspb.Message {
  getAnalysis(): string;
  setAnalysis(value: string): void;

  getDone(): boolean;
  setDone(value: boolean): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AnalyzeNodeResponse.AsObject;
  static toObject(includeInstance: boolean, msg: AnalyzeNodeResponse): AnalyzeNodeResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: AnalyzeNodeResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AnalyzeNodeResponse;
  static deserializeBinaryFromReader(message: AnalyzeNodeResponse, reader: jspb.BinaryReader): AnalyzeNodeResponse;
}

export namespace AnalyzeNodeResponse {
  export type AsObject = {
    analysis: string,
    done: boolean,
  }
}

