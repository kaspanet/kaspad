package blockdag

import (
	"errors"
	"os"
	"strings"
	"testing"

	"bou.ke/monkey"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/database"
)

func TestIsSupportedDbType(t *testing.T) {
	if !isSupportedDbType("ffldb") {
		t.Errorf("ffldb should be a supported DB driver")
	}
	if isSupportedDbType("madeUpDb") {
		t.Errorf("madeUpDb should not be a supported DB driver")
	}
}

// TestDAGSetupErrors tests all error-cases in DAGSetup.
// The non-error-cases are tested in the more general tests.
func TestDAGSetupErrors(t *testing.T) {
	os.RemoveAll(testDbRoot)
	testDAGSetupErrorThroughPatching(t, "unable to create test db root: ", os.MkdirAll, func(path string, perm os.FileMode) error {
		return errors.New("Made up error")
	})

	testDAGSetupErrorThroughPatching(t, "failed to create dag instance: ", New, func(config *Config) (*BlockDAG, error) {
		return nil, errors.New("Made up error")
	})

	testDAGSetupErrorThroughPatching(t, "unsupported db type ", isSupportedDbType, func(dbType string) bool {
		return false
	})

	testDAGSetupErrorThroughPatching(t, "error creating db: ", database.Create, func(dbType string, args ...interface{}) (database.DB, error) {
		return nil, errors.New("Made up error")
	})
}

func testDAGSetupErrorThroughPatching(t *testing.T, expectedErrorMessage string, targetFunction interface{}, replacementFunction interface{}) {
	monkey.Patch(targetFunction, replacementFunction)
	_, tearDown, err := DAGSetup("TestDAGSetup", &dagconfig.MainNetParams)
	if tearDown != nil {
		defer tearDown()
	}
	if err == nil || !strings.HasPrefix(err.Error(), expectedErrorMessage) {
		t.Errorf("DAGSetup: expected error to have prefix '%s' but got error '%v'", expectedErrorMessage, err)
	}
	monkey.Unpatch(targetFunction)
}
