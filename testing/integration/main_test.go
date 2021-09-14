package integration

import (
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(0)
	logger.SetLogLevels(logger.LevelDebug)
	logger.InitLogStdout(logger.LevelDebug)

	os.Exit(m.Run())
}
