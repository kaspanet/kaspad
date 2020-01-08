package blockdag

import (
	"github.com/pkg/errors"
	"strings"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/database"
)

func TestAncestorErrors(t *testing.T) {
	// Create a new database and DAG instance to run tests against.
	params := dagconfig.SimnetParams
	dag, teardownFunc, err := DAGSetup("TestAncestorErrors", Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("TestAncestorErrors: Failed to setup DAG instance: %s", err)
	}
	defer teardownFunc()

	node := newTestNode(dag, newSet(), int32(0x10000000), 0, time.Unix(0, 0))
	node.chainHeight = 2
	ancestor := node.SelectedAncestor(3)
	if ancestor != nil {
		t.Errorf("TestAncestorErrors: Ancestor() unexpectedly returned a node. Expected: <nil>")
	}
}

func TestFlushToDBErrors(t *testing.T) {
	// Create a new database and DAG instance to run tests against.
	dag, teardownFunc, err := DAGSetup("TestFlushToDBErrors", Config{
		DAGParams: &dagconfig.SimnetParams,
	})
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
	guard := monkey.Patch(dbStoreBlockNode, func(_ database.Tx, _ *blockNode) error {
		return errors.New(databaseErrorMessage)
	})
	defer guard.Unpatch()
	err = dag.index.flushToDB()
	if err == nil {
		t.Errorf("TestFlushToDBErrors: flushToDB failure due to database error: "+
			"Expected: %s, got: <nil>", databaseErrorMessage)
	}
	if !strings.Contains(err.Error(), databaseErrorMessage) {
		t.Errorf("TestFlushToDBErrors: flushToDB failure due to database error: "+
			"Unexpected flushToDB error. Expected: %s, got: %s", databaseErrorMessage, err)
	}
}
