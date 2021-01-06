package main

import (
	"fmt"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/util/panics"
	"os"
)

var (
	backendLog = logger.NewBackend()
	log        = backendLog.Logger("KSMN")
	spawn      = panics.GoroutineWrapperFunc(log)
)

func initLog(logFile, errLogFile string) {
	err := backendLog.AddLogFile(logFile, logger.LevelTrace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error adding log file %s as log rotator for level %s: %s", logFile, logger.LevelTrace, err)
		os.Exit(1)
	}
	err = backendLog.AddLogFile(errLogFile, logger.LevelWarn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error adding log file %s as log rotator for level %s: %s", errLogFile, logger.LevelWarn, err)
		os.Exit(1)
	}
}
