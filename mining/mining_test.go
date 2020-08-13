package mining

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/domainmessage"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"testing"
)

func TestIncestousNewBlockTemplate(t *testing.T) {
	// Create a new database and dag instance to run tests against.
	dag, teardownFunc, err := blockdag.DAGSetup("TestChainedTransactions", true, blockdag.Config{
		DAGParams: &dagconfig.SimnetParams,
	})
	if err != nil {
		t.Fatalf("Failed to setup dag instance: %v", err)
	}
	defer teardownFunc()

	// Create a block over genesis but don't submit it
	// Note that even though we're calling PrepareBlockForTest for
	// convenience's sake, what we're actually testing is
	// NewBlockTemplate, which is being called by PrepareBlockForTest.
	heldBlock, err := PrepareBlockForTest(dag, []*daghash.Hash{dag.Params.GenesisHash}, []*domainmessage.MsgTx{}, false)
	if err != nil {
		t.Fatalf("unexpected error in PrepareBlockForTest: %s", err)
	}

	// Add a chain with size `chainSize` over the genesis
	const chainSize = 1010
	chainTipHash := dag.Params.GenesisHash
	for i := 0; i < chainSize; i++ {
		block, err := PrepareBlockForTest(dag, []*daghash.Hash{chainTipHash}, []*domainmessage.MsgTx{}, false)
		if err != nil {
			t.Fatalf("unexpected error in PrepareBlockForTest: %s", err)
		}
		isOrphan, isDelayed, err := dag.ProcessBlock(util.NewBlock(block), blockdag.BFNoPoWCheck)
		if err != nil {
			t.Fatalf("block #%d unexpectedly got an error in ProcessBlock: %s", i, err)
		}
		if isOrphan {
			t.Fatalf("block #%d is unexpectedly an orphan", i)
		}
		if isDelayed {
			t.Fatalf("block #%d is unexpectedly delayed", i)
		}
		chainTipHash = block.BlockHash()
	}

	// Add `heldBlock` to the DAG
	isOrphan, isDelayed, err := dag.ProcessBlock(util.NewBlock(heldBlock), blockdag.BFNoPoWCheck)
	if err != nil {
		t.Fatalf("unexpected error in ProcessBlock: %s", err)
	}
	if isOrphan {
		t.Fatalf("held block is unexpectedly an orphan")
	}
	if isDelayed {
		t.Fatalf("held block is unexpectedly delayed")
	}

	// Create and add a block whose parents are the last chain block and heldBlock.
	// We expect this not to fail.
	block, err := PrepareBlockForTest(dag, []*daghash.Hash{chainTipHash, heldBlock.BlockHash()},
		[]*domainmessage.MsgTx{}, false)
	if err != nil {
		t.Fatalf("unexpected error in PrepareBlockForTest: %s", err)
	}
	isOrphan, isDelayed, err = dag.ProcessBlock(util.NewBlock(block), blockdag.BFNoPoWCheck)
	if err != nil {
		t.Fatalf("unexpected error in ProcessBlock: %s", err)
	}
	if isOrphan {
		t.Fatalf("held block is unexpectedly an orphan")
	}
	if isDelayed {
		t.Fatalf("held block is unexpectedly delayed")
	}
}
