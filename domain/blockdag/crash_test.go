package blockdag

import (
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util/daghash"
	"testing"
)

func TestCrashingDAG(t *testing.T) {
	// Create a new database and DAG instance to run tests against.
	dag, teardownFunc, err := DAGSetup("TestCrashingDAG", true, Config{
		DAGParams: &dagconfig.SimnetParams,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()

	genesis := dag.genesis
	block1 := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{genesis.hash}, nil)
	block2 := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{block1.BlockHash()}, nil)
	block3 := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{block2.BlockHash()}, nil)
	block4 := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{genesis.hash}, nil)
	block5 := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{block3.BlockHash(), block4.BlockHash()}, nil)
	block6 := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{genesis.hash}, nil)
	block7 := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{block6.BlockHash(), block1.BlockHash(), block4.BlockHash()}, nil)
	PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{block5.BlockHash(), block7.BlockHash()}, nil)
	PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{block7.BlockHash(), block2.BlockHash()}, nil)
}
