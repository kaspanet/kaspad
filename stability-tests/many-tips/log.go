package main

import (
	"fmt"
	"github.com/c4ei/kaspad/infrastructure/logger"
	"github.com/c4ei/kaspad/stability-tests/common"
	"github.com/c4ei/kaspad/util/panics"
	"os"
)

var (
	backendLog = logger.NewBackend()
	log        = backendLog.Logger("MATS")
	spawn      = panics.GoroutineWrapperFunc(log)
)

func initLog(logFile, errLogFile string) {
	level := logger.LevelDebug
	if activeConfig().LogLevel != "" {
		var ok bool
		level, ok = logger.LevelFromString(activeConfig().LogLevel)
		if !ok {
			fmt.Fprintf(os.Stderr, "Log level %s doesn't exists", activeConfig().LogLevel)
			os.Exit(1)
		}
	}
	log.SetLevel(level)
	common.InitBackend(backendLog, logFile, errLogFile)
}
