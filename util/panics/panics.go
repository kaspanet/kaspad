package panics

import (
	"os"
	"runtime/debug"

	"github.com/btcsuite/btclog"
	"github.com/daglabs/btcd/logger"
)

// HandlePanic recovers panics, log them, and then exits the process.
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

// GoroutineWrapperFunc returns a goroutine wrapper function that handles panics and write them to the log.
func GoroutineWrapperFunc(log btclog.Logger) func(func()) {
	return func(f func()) {
		go func() {
			defer HandlePanic(log)
			f()
		}()
	}
}
