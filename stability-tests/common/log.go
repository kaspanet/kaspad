package common

import (
	"fmt"
	"os"

	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/stability-tests/common/mine"
	"github.com/kaspanet/kaspad/stability-tests/common/rpc"
	"github.com/kaspanet/kaspad/util/panics"
)

// log is a logger that is initialized with no output filters. This
// means the package will not perform any logging by default until the caller
// requests it.
var log *logger.Logger
var spawn func(name string, spawnedFunction func())

const logSubsytem = "STCM"

// The default amount of logging is none.
func init() {
	DisableLog()
}

// DisableLog disables all library log output. Logging output is disabled
// by default until UseLogger is called.
func DisableLog() {
	backend := logger.NewBackend()
	log = backend.Logger(logSubsytem)
	log.SetLevel(logger.LevelOff)
	spawn = panics.GoroutineWrapperFunc(log)
	logger.SetLogLevels(logger.LevelOff)
	logger.InitLogStdout(logger.LevelInfo)
}

// UseLogger uses a specified Logger to output package logging info.
func UseLogger(backend *logger.Backend, level logger.Level) {
	log = backend.Logger(logSubsytem)
	log.SetLevel(level)
	spawn = panics.GoroutineWrapperFunc(log)

	mine.UseLogger(backend, level)
	rpc.UseLogger(backend, level)
	logger.SetLogLevels(level)
}

func InitBackend(backendLog *logger.Backend, logFile, errLogFile string) {
	err := backendLog.AddLogFile(logFile, logger.LevelTrace)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error adding log file %s as log rotator for level %s: %+v\n", logFile, logger.LevelTrace, err)
		os.Exit(1)
	}
	err = backendLog.AddLogFile(errLogFile, logger.LevelWarn)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error adding log file %s as log rotator for level %s: %+v\n", errLogFile, logger.LevelWarn, err)
		os.Exit(1)
	}

	err = backendLog.AddLogWriter(os.Stdout, logger.LevelDebug)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error adding stdout to the loggerfor level %s: %+v\n", logger.LevelInfo, err)
		os.Exit(1)
	}

	err = backendLog.Run()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error starting the logger: %s ", err)
		os.Exit(1)
	}
}
