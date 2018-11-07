package ffldb

import (
	"errors"
	"testing"

	"bou.ke/monkey"
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
