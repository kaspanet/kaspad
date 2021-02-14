package integration

import (
	"os"
	"testing"

	"github.com/kaspanet/kaspad/infrastructure/logger"
)

func TestMain(m *testing.M) {
	logger.SetLogLevels(logger.LevelDebug)
	logger.InitLogStdoutOnly(logger.LevelDebug)

	os.Exit(m.Run())
}
