package mempoollimits

import (
	"fmt"
	"os"

	"github.com/c4ei/YunSeokYeol/infrastructure/logger"
	"github.com/c4ei/YunSeokYeol/stability-tests/common"
	"github.com/c4ei/YunSeokYeol/util/panics"
)

var (
	backendLog = logger.NewBackend()
	log        = backendLog.Logger("MPLM")
	spawn      = panics.GoroutineWrapperFunc(log)
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
