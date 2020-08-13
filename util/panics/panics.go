package panics

import (
	"fmt"
	"os"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/kaspanet/kaspad/infrastructure/logs"
)

const exitHandlerTimeout = 5 * time.Second

// HandlePanic recovers panics and then initiates a clean shutdown.
func HandlePanic(log *logs.Logger, goroutineName string, goroutineStackTrace []byte) {
	err := recover()
	if err == nil {
		return
	}

	reason := fmt.Sprintf("Fatal error in goroutine `%s`: %+v", goroutineName, err)
	exit(log, reason, debug.Stack(), goroutineStackTrace)
}

var goroutineLastID uint64

// GoroutineWrapperFunc returns a goroutine wrapper function that handles panics and writes them to the log.
func GoroutineWrapperFunc(log *logs.Logger) func(name string, spawnedFunction func()) {
	return func(name string, f func()) {
		stackTrace := debug.Stack()
		go func() {
			handleSpawnedFunction(log, stackTrace, name, f)
		}()
	}
}

// AfterFuncWrapperFunc returns a time.AfterFunc wrapper function that handles panics.
func AfterFuncWrapperFunc(log *logs.Logger) func(name string, d time.Duration, f func()) *time.Timer {
	return func(name string, d time.Duration, f func()) *time.Timer {
		stackTrace := debug.Stack()
		return time.AfterFunc(d, func() {
			handleSpawnedFunction(log, stackTrace, name, f)
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

func handleSpawnedFunction(log *logs.Logger, stackTrace []byte, spawnedFunctionName string, spawnedFunction func()) {
	goroutineID := atomic.AddUint64(&goroutineLastID, 1)
	goroutineName := fmt.Sprintf("%s %d", spawnedFunctionName, goroutineID)
	utilLog.Debugf("Started goroutine `%s`", goroutineName)
	defer utilLog.Debugf("Ended goroutine `%s`", goroutineName)
	defer HandlePanic(log, goroutineName, stackTrace)
	spawnedFunction()
}
