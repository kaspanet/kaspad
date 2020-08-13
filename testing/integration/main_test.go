package integration

import (
	"os"
	"testing"

	"github.com/kaspanet/kaspad/infrastructure/logger"
)

func TestMain(m *testing.M) {
	logger.SetLogLevels("debug")

	os.Exit(m.Run())
}
