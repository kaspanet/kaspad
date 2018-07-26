package database

import (
	"os"
	"testing"

	"github.com/btcsuite/btclog"
)

func TestUseLogger(t *testing.T) {
	testLogger := btclog.NewBackend(os.Stdout)
	defer os.Stdout.Sync()
	dbLog := testLogger.Logger("BCDB")
	dbLog.SetLevel(btclog.LevelOff)
	UseLogger(dbLog)

	useLoggerCalled := false
	useLoggerFunc := func(logger btclog.Logger) {
		useLoggerCalled = true
		if logger != dbLog {
			t.Errorf("TestUseLogger: driver's logger is expected to be dbLog, but it is not")
		}
	}

	driver := Driver{
		UseLogger: useLoggerFunc,
	}
	RegisterDriver(driver)

	UseLogger(dbLog)

	if log != dbLog {
		t.Errorf("TestUseLogger: Log is expected to be dbLog but it is not")
	}

	if !useLoggerCalled {
		t.Errorf("TestUseLogger: driver.UseLogger was not called")
	}
}
