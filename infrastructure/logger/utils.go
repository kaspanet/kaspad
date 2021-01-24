package logger

import (
	"runtime"
	"time"
)

// LogAndMeasureExecutionTime logs that `functionName` has
// started. The user is expected to defer `onEnd`, which
// will then log that the function has ended, as well as
// the time duration the function had ran.
func LogAndMeasureExecutionTime(log *Logger, functionName string) (onEnd func()) {
	start := time.Now()
	log.Debugf("%s start", functionName)
	return func() {
		log.Debugf("%s end. Took: %s", functionName, time.Since(start))
	}
}

// LogMemoryStats logs memory stats for `functionName`
func LogMemoryStats(log *Logger, functionName string) {
	stats := runtime.MemStats{}
	runtime.ReadMemStats(&stats)
	log.Debugf("%s: used memory: %d bytes, total: %d bytes", functionName,
		stats.Alloc, stats.HeapIdle-stats.HeapReleased+stats.HeapInuse)
}
