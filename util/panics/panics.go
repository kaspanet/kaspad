package panics

import (
	"os"
	"runtime/debug"

	"github.com/btcsuite/btclog"
	"github.com/daglabs/btcd/logger"
)

func HandlePanic(log btclog.Logger) {
	if err := recover(); err != nil {
		log.Criticalf("Fatal error: %s", err)
		log.Criticalf("Stack trace: %s", debug.Stack())
		if logger.LogRotator != nil {
			logger.LogRotator.Close()
		}
		os.Exit(1)
	}
}
