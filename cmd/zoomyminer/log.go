package main

import (
	"fmt"
	"os"

	"github.com/zoomy-network/zoomyd/infrastructure/logger"
	"github.com/zoomy-network/zoomyd/util/panics"
)

var (
	backendLog = logger.NewBackend()
	log        = backendLog.Logger("KSMN")
	spawn      = panics.GoroutineWrapperFunc(log)
)

func initLog(logFile, errLogFile string) {
	log.SetLevel(logger.LevelDebug)
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
	err = backendLog.AddLogWriter(os.Stdout, logger.LevelInfo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error adding stdout to the loggerfor level %s: %s", logger.LevelWarn, err)
		os.Exit(1)
	}
	err = backendLog.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting the logger: %s ", err)
		os.Exit(1)
	}

}
