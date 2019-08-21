package panics

import (
	"github.com/daglabs/btcd/logs"
	"os"
	"runtime/debug"
)

// HandlePanic recovers panics, log them, and then exits the process.
func HandlePanic(log logs.Logger, backendLog *logs.Backend) {
	if err := recover(); err != nil {
		log.Criticalf("Fatal error: %s", err)
		log.Criticalf("Stack trace: %s", debug.Stack())
		if backendLog != nil {
			backendLog.Close()
		}
		os.Exit(1)
	}
}

// GoroutineWrapperFunc returns a goroutine wrapper function that handles panics and write them to the log.
func GoroutineWrapperFunc(log logs.Logger, backendLog *logs.Backend) func(func()) {
	return func(f func()) {
		go func() {
			defer HandlePanic(log, backendLog)
			f()
		}()
	}
}
