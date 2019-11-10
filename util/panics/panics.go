package panics

import (
	"github.com/daglabs/btcd/logs"
	"os"
	"runtime/debug"
)

// HandlePanic recovers panics, log them, and then exits the process.
func HandlePanic(log logs.Logger, backendLog *logs.Backend, goroutineStackTrace []byte) {
	if err := recover(); err != nil {
		log.Criticalf("Fatal error: %+v", err)
		log.Criticalf("goroutine stack trance: %s", goroutineStackTrace)
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
		stackTrace := debug.Stack()
		go func() {
			defer HandlePanic(log, backendLog, stackTrace)
			f()
		}()
	}
}
