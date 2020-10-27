package logger

import (
	"time"
)

func LogAndMeasureExecutionTime(log *Logger, functionName string) (onEnd func()) {
	start := time.Now()
	log.Debugf("%s start", functionName)
	return func() {
		log.Debugf("%s end. Took: %s", functionName, time.Since(start))
	}
}
