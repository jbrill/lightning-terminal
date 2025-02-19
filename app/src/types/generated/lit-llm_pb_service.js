// package: litrpc
// file: lit-llm.proto

var lit_llm_pb = require("./lit-llm_pb");
var grpc = require("@improbable-eng/grpc-web").grpc;

var LLM = (function () {
  function LLM() {}
  LLM.serviceName = "litrpc.LLM";
  return LLM;
}());

LLM.AnalyzeNode = {
  methodName: "AnalyzeNode",
  service: LLM,
  requestStream: false,
  responseStream: true,
  requestType: lit_llm_pb.AnalyzeNodeRequest,
  responseType: lit_llm_pb.AnalyzeNodeResponse
};

exports.LLM = LLM;

function LLMClient(serviceHost, options) {
  this.serviceHost = serviceHost;
  this.options = options || {};
}

LLMClient.prototype.analyzeNode = function analyzeNode(requestMessage, metadata) {
  var listeners = {
    data: [],
    end: [],
    status: []
  };
  var client = grpc.invoke(LLM.AnalyzeNode, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onMessage: function (responseMessage) {
      listeners.data.forEach(function (handler) {
        handler(responseMessage);
      });
    },
    onEnd: function (status, statusMessage, trailers) {
      listeners.status.forEach(function (handler) {
        handler({ code: status, details: statusMessage, metadata: trailers });
      });
      listeners.end.forEach(function (handler) {
        handler({ code: status, details: statusMessage, metadata: trailers });
      });
      listeners = null;
    }
  });
  return {
    on: function (type, handler) {
      listeners[type].push(handler);
      return this;
    },
    cancel: function () {
      listeners = null;
      client.close();
    }
  };
};

exports.LLMClient = LLMClient;

