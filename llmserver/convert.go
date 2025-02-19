package llmserver

import (
	"github.com/lightninglabs/lightning-terminal/litrpc"
)

// ConvertAnalysis converts an internal Analysis to an RPC response.
func ConvertAnalysis(a *Analysis) *litrpc.AnalyzeNodeResponse {
	return &litrpc.AnalyzeNodeResponse{
		Analysis: a.Content,
	}
}
