package logger

import (
	"fmt"
	"github.com/daglabs/btcd/logs"
	"os"
)

// BackendLog is the logging backend used to create all subsystem loggers.
var BackendLog = logs.NewBackend()

// InitLog attaches log file and error log file to the backend log.
func InitLog(logFile, errLogFile string) {
	err := BackendLog.AddLogFile(logFile, logs.LevelTrace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error adding log file %s as log rotator for level %s: %s", logFile, logs.LevelTrace, err)
		os.Exit(1)
	}
	err = BackendLog.AddLogFile(errLogFile, logs.LevelWarn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error adding log file %s as log rotator for level %s: %s", errLogFile, logs.LevelWarn, err)
		os.Exit(1)
	}
}
