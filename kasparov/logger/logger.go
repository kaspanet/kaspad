package logger

import (
	"fmt"
	"github.com/daglabs/kaspad/logs"
	"github.com/pkg/errors"
	"os"
)

// BackendLog is the logging backend used to create all subsystem loggers.
var BackendLog = logs.NewBackend()
var loggers []logs.Logger

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

// Logger returns a new logger for a particular subsystem that writes to
// BackendLog, and add it to a slice so it will be possible to access it
// later and change its log level
func Logger(subsystemTag string) logs.Logger {
	logger := BackendLog.Logger(subsystemTag)
	loggers = append(loggers, logger)
	return logger
}

// SetLogLevels sets the logging level for all of the subsystems in Kasparov.
func SetLogLevels(level string) error {
	lvl, ok := logs.LevelFromString(level)
	if !ok {
		return errors.Errorf("Invalid log level %s", level)
	}
	for _, logger := range loggers {
		logger.SetLevel(lvl)
	}
	return nil
}
