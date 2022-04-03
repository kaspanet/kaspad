package logger

import (
	"fmt"
	"runtime"
	"time"
)

// LogAndMeasureExecutionTime logs that `functionName` has
// started. The user is expected to defer `onEnd`, which
// will then log that the function has ended, as well as
// the time duration the function had ran.
func LogAndMeasureExecutionTime(log *Logger, functionName string) (onEnd func()) {
	start := time.Now()
	log.Tracef("%s start", functionName)
	return func() {
		log.Tracef("%s end. Took: %s", functionName, time.Since(start))
	}
}

// LogMemoryStats logs memory stats for `functionName`
func LogMemoryStats(log *Logger, functionName string) {
	log.Debug(NewLogClosure(func() string {
		stats := runtime.MemStats{}
		runtime.ReadMemStats(&stats)
		return fmt.Sprintf("%s: used memory: %d bytes, total: %d bytes", functionName,
			stats.Alloc, stats.HeapIdle-stats.HeapReleased+stats.HeapInuse)
	}))
}

// LogClosure is a closure that can be printed with %s to be used to
// generate expensive-to-create data for a detailed log level and avoid doing
// the work if the data isn't printed.
type LogClosure func() string

func (c LogClosure) String() string {
	return c()
}

// NewLogClosure casts a function to a LogClosure.
// See LogClosure for details.
func NewLogClosure(c func() string) LogClosure {
	return c
}
