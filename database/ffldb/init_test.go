package ffldb

import (
	"errors"
	"os"
	"testing"

	"github.com/bouk/monkey"
	"github.com/btcsuite/btclog"
	"github.com/daglabs/btcd/database"
)

// TestRegisterDriverErrors tests all error-cases in registerDriver().
// The non-error-cases are tested in the more general tests.
func TestInitErrors(t *testing.T) {
	patch := monkey.Patch(database.RegisterDriver,
		func(driver database.Driver) error { return errors.New("Error in database.RegisterDriver") })
	defer patch.Unpatch()

	defer func() {
		err := recover()
		if err == nil {
			t.Errorf("TestRegisterDriverErrors: No panic on init when database.RegisterDriver returned an error")
		}
	}()

	registerDriver()
}

func TestUseLogger(t *testing.T) {
	currentLog := log
	defer func() { log = currentLog }()

	backend := btclog.NewBackend(os.Stdout)
	logger := backend.Logger("TEST")

	useLogger(logger)

	if log != logger {
		t.Errorf("TestUseLogger: `log` was not changed to correct logger")
	}

}
