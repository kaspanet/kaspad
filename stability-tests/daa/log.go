package daa

import (
	"fmt"
	"os"

	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/stability-tests/common"
)

var (
	backendLog = logger.NewBackend()
	log        = backendLog.Logger("DAAT")
)

func initLog(logFile, errLogFile string) {
	level := logger.LevelInfo
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
