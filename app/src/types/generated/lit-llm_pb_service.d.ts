// package: litrpc
// file: lit-llm.proto

import * as lit_llm_pb from "./lit-llm_pb";
import {grpc} from "@improbable-eng/grpc-web";

type LLMAnalyzeNode = {
  readonly methodName: string;
  readonly service: typeof LLM;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof lit_llm_pb.AnalyzeNodeRequest;
  readonly responseType: typeof lit_llm_pb.AnalyzeNodeResponse;
};

export class LLM {
  static readonly serviceName: string;
  static readonly AnalyzeNode: LLMAnalyzeNode;
}

export type ServiceError = { message: string, code: number; metadata: grpc.Metadata }
export type Status = { details: string, code: number; metadata: grpc.Metadata }

interface UnaryResponse {
  cancel(): void;
}
interface ResponseStream<T> {
  cancel(): void;
  on(type: 'data', handler: (message: T) => void): ResponseStream<T>;
  on(type: 'end', handler: (status?: Status) => void): ResponseStream<T>;
  on(type: 'status', handler: (status: Status) => void): ResponseStream<T>;
}
interface RequestStream<T> {
  write(message: T): RequestStream<T>;
  end(): void;
  cancel(): void;
  on(type: 'end', handler: (status?: Status) => void): RequestStream<T>;
  on(type: 'status', handler: (status: Status) => void): RequestStream<T>;
}
interface BidirectionalStream<ReqT, ResT> {
  write(message: ReqT): BidirectionalStream<ReqT, ResT>;
  end(): void;
  cancel(): void;
  on(type: 'data', handler: (message: ResT) => void): BidirectionalStream<ReqT, ResT>;
  on(type: 'end', handler: (status?: Status) => void): BidirectionalStream<ReqT, ResT>;
  on(type: 'status', handler: (status: Status) => void): BidirectionalStream<ReqT, ResT>;
}

export class LLMClient {
  readonly serviceHost: string;

  constructor(serviceHost: string, options?: grpc.RpcOptions);
  analyzeNode(
    requestMessage: lit_llm_pb.AnalyzeNodeRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: lit_llm_pb.AnalyzeNodeResponse|null) => void
  ): UnaryResponse;
  analyzeNode(
    requestMessage: lit_llm_pb.AnalyzeNodeRequest,
    callback: (error: ServiceError|null, responseMessage: lit_llm_pb.AnalyzeNodeResponse|null) => void
  ): UnaryResponse;
}

