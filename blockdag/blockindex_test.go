package blockdag

import (
	"testing"
	"time"

	"github.com/kaspanet/kaspad/dagconfig"
)

func TestAncestorErrors(t *testing.T) {
	// Create a new database and DAG instance to run tests against.
	params := dagconfig.SimnetParams
	dag, teardownFunc, err := DAGSetup("TestAncestorErrors", true, Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("TestAncestorErrors: Failed to setup DAG instance: %s", err)
	}
	defer teardownFunc()

	node := newTestNode(dag, newBlockSet(), int32(0x10000000), 0, time.Unix(0, 0))
	node.blueScore = 2
	ancestor := node.SelectedAncestor(3)
	if ancestor != nil {
		t.Errorf("TestAncestorErrors: Ancestor() unexpectedly returned a node. Expected: <nil>")
	}
}
