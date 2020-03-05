package panics

import (
	"fmt"
	"github.com/kaspanet/kaspad/logs"
	"os"
	"runtime/debug"
	"time"
)

// HandlePanic recovers panics, log them, runs an optional panicHandler,
// and then initiates a clean shutdown.
func HandlePanic(backend *logs.Backend, goroutineStackTrace []byte) {
	const panicSubsystem = "PANC"

	log := backend.Logger(panicSubsystem)
	err := recover()
	if err == nil {
		return
	}

	panicHandlerDone := make(chan struct{})
	go func() {
		log.Criticalf("Fatal error: %+v", err)
		if goroutineStackTrace != nil {
			log.Criticalf("Goroutine stack trace: %s", goroutineStackTrace)
		}
		log.Criticalf("Stack trace: %s", debug.Stack())
		backend.Close()
		panicHandlerDone <- struct{}{}
	}()

	const panicHandlerTimeout = 5 * time.Second
	select {
	case <-time.Tick(panicHandlerTimeout):
		fmt.Fprintln(os.Stderr, "Couldn't handle a fatal error. Exiting...")
	case <-panicHandlerDone:
	}
	os.Exit(1)
}

// GoroutineWrapperFunc returns a goroutine wrapper function that handles panics and writes them to the log.
func GoroutineWrapperFunc(backend *logs.Backend) func(func()) {
	return func(f func()) {
		stackTrace := debug.Stack()
		go func() {
			defer HandlePanic(backend, stackTrace)
			f()
		}()
	}
}

// AfterFuncWrapperFunc returns a time.AfterFunc wrapper function that handles panics.
func AfterFuncWrapperFunc(backend *logs.Backend) func(d time.Duration, f func()) *time.Timer {
	return func(d time.Duration, f func()) *time.Timer {
		stackTrace := debug.Stack()
		return time.AfterFunc(d, func() {
			defer HandlePanic(backend, stackTrace)
			f()
		})
	}
}
