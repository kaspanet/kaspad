package reachabilitymanager_test

import (
	"os"
	"testing"

	"github.com/kaspanet/kaspad/infrastructure/logger"
)

const logLevel = logger.LevelWarn

func TestMain(m *testing.M) {
	logger.SetLogLevels(logLevel)
	logger.InitLogStdoutOnly(logLevel)
	os.Exit(m.Run())
}
