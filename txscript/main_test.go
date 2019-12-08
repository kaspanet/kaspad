package txscript

import (
	"os"
	"testing"

	"github.com/kaspanet/kaspad/logs"
)

func TestMain(m *testing.M) {
	// set log level to trace, so that logClosures passed to log.Tracef are covered
	log.SetLevel(logs.LevelTrace)

	os.Exit(m.Run())
}
