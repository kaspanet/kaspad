package panics

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/kaspanet/kaspad/logs"
)

const exitHandlerTimeout = 5 * time.Second

// HandlePanic recovers panics and then initiates a clean shutdown.
func HandlePanic(log *logs.Logger, goroutineStackTrace []byte) {
	err := recover()
	if err == nil {
		return
	}

	reason := fmt.Sprintf("Fatal error: %+v", err)
	exit(log, reason, debug.Stack(), goroutineStackTrace)
}

// GoroutineWrapperFunc returns a goroutine wrapper function that handles panics and writes them to the log.
func GoroutineWrapperFunc(log *logs.Logger) func(func()) {
	return func(f func()) {
		stackTrace := debug.Stack()
		go func() {
			defer HandlePanic(log, stackTrace)
			f()
		}()
	}
}

// AfterFuncWrapperFunc returns a time.AfterFunc wrapper function that handles panics.
func AfterFuncWrapperFunc(log *logs.Logger) func(d time.Duration, f func()) *time.Timer {
	return func(d time.Duration, f func()) *time.Timer {
		stackTrace := debug.Stack()
		return time.AfterFunc(d, func() {
			defer HandlePanic(log, stackTrace)
			f()
		})
	}
}

// Exit prints the given reason to log and initiates a clean shutdown.
func Exit(log *logs.Logger, reason string) {
	exit(log, reason, nil, nil)
}

// Exit prints the given reason, prints either of the given stack traces (if not nil),
// waits for them to finish writing, and exits.
func exit(log *logs.Logger, reason string, currentThreadStackTrace []byte, goroutineStackTrace []byte) {
	exitHandlerDone := make(chan struct{})
	go func() {
		log.Criticalf("Exiting: %s", reason)
		if goroutineStackTrace != nil {
			log.Criticalf("Goroutine stack trace: %s", goroutineStackTrace)
		}
		if currentThreadStackTrace != nil {
			log.Criticalf("Stack trace: %s", currentThreadStackTrace)
		}
		log.Backend().Close()
		close(exitHandlerDone)
	}()

	select {
	case <-time.After(exitHandlerTimeout):
		fmt.Fprintln(os.Stderr, "Couldn't exit gracefully.")
	case <-exitHandlerDone:
	}
	fmt.Print("Exiting...")
	os.Exit(1)
	fmt.Print("After os.Exit(1)")
}
