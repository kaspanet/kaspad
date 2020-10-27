package logger

import (
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
