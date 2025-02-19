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
  responseStream: false,
  requestType: lit_llm_pb.AnalyzeNodeRequest,
  responseType: lit_llm_pb.AnalyzeNodeResponse
};

exports.LLM = LLM;

function LLMClient(serviceHost, options) {
  this.serviceHost = serviceHost;
  this.options = options || {};
}

LLMClient.prototype.analyzeNode = function analyzeNode(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(LLM.AnalyzeNode, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

exports.LLMClient = LLMClient;

