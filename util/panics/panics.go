package panics

import (
	"github.com/kaspanet/kaspad/logs"
	"github.com/kaspanet/kaspad/signal"
	"runtime/debug"
	"time"
)

// HandlePanic recovers panics, log them, runs an optional panicHandler,
// and then initiates a clean shutdown.
func HandlePanic(log logs.Logger, goroutineStackTrace []byte, panicHandler func()) {
	if err := recover(); err != nil {
		log.Criticalf("Fatal error: %+v", err)
		if goroutineStackTrace != nil {
			log.Criticalf("goroutine stack trace: %s", goroutineStackTrace)
		}
		log.Criticalf("Stack trace: %s", debug.Stack())
		if panicHandler != nil {
			panicHandler()
		}
		signal.PanicShutdownChannel <- struct{}{}
	}
}

// GoroutineWrapperFunc returns a goroutine wrapper function that handles panics and writes them to the log.
func GoroutineWrapperFunc(log logs.Logger) func(func()) {
	return func(f func()) {
		stackTrace := debug.Stack()
		go func() {
			defer HandlePanic(log, stackTrace, nil)
			f()
		}()
	}
}

// GoroutineWrapperFuncWithPanicHandler returns a goroutine wrapper function that handles panics,
// writes them to the log, and executes a handler function for panics.
func GoroutineWrapperFuncWithPanicHandler(log logs.Logger) func(func(), func()) {
	return func(f func(), panicHandler func()) {
		stackTrace := debug.Stack()
		go func() {
			defer HandlePanic(log, stackTrace, panicHandler)
			f()
		}()
	}
}

// AfterFuncWrapperFunc returns a time.AfterFunc wrapper function that handles panics,
func AfterFuncWrapperFunc(log logs.Logger) func(d time.Duration, f func()) *time.Timer {
	return func(d time.Duration, f func()) *time.Timer {
		stackTrace := debug.Stack()
		return time.AfterFunc(d, func() {
			defer HandlePanic(log, stackTrace, nil)
			f()
		})
	}
}

// AfterFuncWrapperFuncWithPanicHandler returns a time.AfterFunc wrapper function that handles panics,
// writes them to the log, and executes a handler function for panics.
func AfterFuncWrapperFuncWithPanicHandler(log logs.Logger) func(d time.Duration, f func(), panicHandler func()) *time.Timer {
	return func(d time.Duration, f func(), panicHandler func()) *time.Timer {
		stackTrace := debug.Stack()
		return time.AfterFunc(d, func() {
			defer HandlePanic(log, stackTrace, panicHandler)
			f()
		})
	}
}
