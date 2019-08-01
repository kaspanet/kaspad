package panics

import (
	"github.com/daglabs/btcd/logs"
	"os"
	"runtime/debug"

	"github.com/daglabs/btcd/logger"
)

// HandlePanic recovers panics, log them, and then exits the process.
func HandlePanic(log logs.Logger) {
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
func GoroutineWrapperFunc(log logs.Logger) func(func()) {
	return func(f func()) {
		go func() {
			defer HandlePanic(log)
			f()
		}()
	}
}
