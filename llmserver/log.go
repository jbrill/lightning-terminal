// llm/log.go
package llmserver

import (
    "github.com/btcsuite/btclog"
    "github.com/lightningnetwork/lnd/build"
)

const Subsystem = "LLM"

var log btclog.Logger

func init() {
    UseLogger(build.NewSubLogger(Subsystem, nil))
}

func UseLogger(logger btclog.Logger) {
    log = logger
}
