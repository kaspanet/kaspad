package integration

import (
	"os"
	"testing"

	"github.com/kaspanet/kaspad/logger"
)

func TestMain(m *testing.M) {
	logger.SetLogLevels("trace") // TODO: Set back to debug

	os.Exit(m.Run())
}
