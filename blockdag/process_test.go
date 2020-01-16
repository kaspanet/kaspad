package blockdag

import (
	"path/filepath"
	"testing"

	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/util/daghash"
)

func TestProcessOrphans(t *testing.T) {
	dag, teardownFunc, err := DAGSetup("TestProcessOrphans", Config{
		DAGParams: &dagconfig.SimnetParams,
	})
	if err != nil {
		t.Errorf("Failed to setup dag instance: %v", err)
		return
	}
	defer teardownFunc()

	dag.TestSetCoinbaseMaturity(0)

	blocksFile := "blk_0_to_4.dat"
	blocks, err := LoadBlocks(filepath.Join("testdata/", blocksFile))
	if err != nil {
		t.Fatalf("TestProcessOrphans: "+
			"Error loading file '%s': %s\n", blocksFile, err)
	}

	// Get a reference to a parent block
	parentBlock := blocks[1]

	// Get a reference to a child block and mess with it so that:
	// a. It gets added to the orphan pool
	// b. It gets rejected once it's unorphaned
	childBlock := blocks[2]
	childBlock.MsgBlock().Header.UTXOCommitment = &daghash.ZeroHash

	// Process the child block so that it gets added to the orphan pool
	isOrphan, isDelayed, err := dag.ProcessBlock(childBlock, BFNoPoWCheck)
	if err != nil {
		t.Fatalf("TestProcessOrphans: child block unexpectedly returned an error: %s", err)
	}
	if isDelayed {
		t.Fatalf("TestProcessOrphans: child block is too far in the future")
	}
	if !isOrphan {
		t.Fatalf("TestProcessOrphans: incorrectly returned that child block is not an orphan")
	}

	// Process the parent block. Note that this will attempt to unorphan the child block
	isOrphan, isDelayed, err = dag.ProcessBlock(parentBlock, BFNone)
	if err != nil {
		t.Fatalf("TestProcessOrphans: parent block unexpectedly returned an error: %s", err)
	}
	if isDelayed {
		t.Fatalf("TestProcessOrphans: parent block is too far in the future")
	}
	if isOrphan {
		t.Fatalf("TestProcessOrphans: incorrectly returned that parent block is an orphan")
	}

	// Make sure that the child block had been rejected
	node := dag.index.LookupNode(childBlock.Hash())
	if node == nil {
		t.Fatalf("TestProcessOrphans: child block missing from block index")
	}
	if !dag.index.NodeStatus(node).KnownInvalid() {
		t.Fatalf("TestProcessOrphans: child block erroneously not marked as invalid")
	}
}
