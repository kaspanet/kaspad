package txscript

import (
	"os"
	"testing"

	"github.com/btcsuite/btclog"
)

func TestMain(m *testing.M) {
	// set log level to trace, so that logClosures passed to log.Tracef are covered
	log.SetLevel(btclog.LevelTrace)

	os.Exit(m.Run())
}
