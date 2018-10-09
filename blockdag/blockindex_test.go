package blockdag

import (
	"strings"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/database"
	"github.com/pkg/errors"
)

func TestAncestorErrors(t *testing.T) {
	node := newTestNode(newSet(), int32(0x10000000), 0, time.Unix(0, 0), dagconfig.MainNetParams.K)
	node.height = 2
	ancestor := node.Ancestor(3)
	if ancestor != nil {
		t.Errorf("TestAncestorErrors: Ancestor() unexpectedly returned a node. Expected: <nil>")
	}
}

func TestFlushToDBErrors(t *testing.T) {
	// Create a new database and DAG instance to run tests against.
	dag, teardownFunc, err := DAGSetup("TestFlushToDBErrors", &dagconfig.MainNetParams)
	if err != nil {
		t.Fatalf("TestFlushToDBErrors: Failed to setup DAG instance: %s", err)
	}
	defer teardownFunc()

	// Call flushToDB without anything to flush. This should succeed
	err = dag.index.flushToDB()
	if err != nil {
		t.Errorf("TestFlushToDBErrors: flushToDB without anything to flush: "+
			"Unexpected flushToDB error: %s", err)
	}

	// Mark the genesis block as dirty
	dag.index.SetStatusFlags(dag.genesis, statusValid)

	// Test flushToDB failure due to database error
	databaseErrorMessage := "database error"
	monkey.Patch(dbStoreBlockNode, func(_ database.Tx, _ *blockNode) error {
		return errors.New(databaseErrorMessage)
	})
	err = dag.index.flushToDB()
	if err == nil {
		t.Errorf("TestFlushToDBErrors: flushToDB failure due to database error: "+
			"Expected: %s, got: <nil>", databaseErrorMessage)
	}
	if !strings.Contains(err.Error(), databaseErrorMessage) {
		t.Errorf("TestFlushToDBErrors: flushToDB failure due to database error: "+
			"Unexpected flushToDB error. Expected: %s, got: %s", databaseErrorMessage, err)
	}
	monkey.Unpatch(dbStoreBlockNode)
}
