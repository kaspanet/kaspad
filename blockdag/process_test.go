package blockdag

import (
	"bou.ke/monkey"
	"fmt"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/daghash"
	"path/filepath"
	"testing"
	"time"
)

func TestProcessBlock(t *testing.T) {
	dag, teardownFunc, err := DAGSetup("TestProcessBlock", Config{
		DAGParams: &dagconfig.SimNetParams,
	})
	if err != nil {
		t.Errorf("Failed to setup dag instance: %v", err)
		return
	}
	defer teardownFunc()

	// Check that BFAfterDelay skip checkBlockSanity
	called := false
	guard := monkey.Patch((*BlockDAG).checkBlockSanity, func(_ *BlockDAG, _ *util.Block, _ BehaviorFlags) (time.Duration, error) {
		called = true
		return 0, nil
	})
	defer guard.Unpatch()

	isOrphan, delay, err := dag.ProcessBlock(util.NewBlock(&Block100000), BFNoPoWCheck)
	if err != nil {
		t.Errorf("ProcessBlock: %s", err)
	}
	if delay != 0 {
		t.Errorf("ProcessBlock: block is too far in the future")
	}
	if !isOrphan {
		t.Errorf("ProcessBlock: unexpected returned non orphan block")
	}
	if !called {
		t.Errorf("ProcessBlock: expected checkBlockSanity to be called")
	}

	Block100000Copy := Block100000
	// Change nonce to change block hash
	Block100000Copy.Header.Nonce++
	called = false
	isOrphan, delay, err = dag.ProcessBlock(util.NewBlock(&Block100000Copy), BFAfterDelay|BFNoPoWCheck)
	if err != nil {
		t.Errorf("ProcessBlock: %s", err)
	}
	if delay != 0 {
		t.Errorf("ProcessBlock: block is too far in the future")
	}
	if !isOrphan {
		t.Errorf("ProcessBlock: unexpected returned non orphan block")
	}
	if called {
		t.Errorf("ProcessBlock: Didn't expected checkBlockSanity to be called")
	}

	isOrphan, delay, err = dag.ProcessBlock(util.NewBlock(dagconfig.SimNetParams.GenesisBlock), BFNone)
	expectedErrMsg := fmt.Sprintf("already have block %s", dagconfig.SimNetParams.GenesisHash)
	if err == nil || err.Error() != expectedErrMsg {
		t.Errorf("ProcessBlock: Expected error \"%s\" but got \"%s\"", expectedErrMsg, err)
	}
}

func TestProcessOrphans(t *testing.T) {
	dag, teardownFunc, err := DAGSetup("TestProcessOrphans", Config{
		DAGParams: &dagconfig.SimNetParams,
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
	isOrphan, delay, err := dag.ProcessBlock(childBlock, BFNoPoWCheck)
	if err != nil {
		t.Fatalf("TestProcessOrphans: child block unexpectedly returned an error: %s", err)
	}
	if delay != 0 {
		t.Fatalf("TestProcessOrphans: child block is too far in the future")
	}
	if !isOrphan {
		t.Fatalf("TestProcessOrphans: incorrectly returned that child block is not an orphan")
	}

	// Process the parent block. Note that this will attempt to unorphan the child block
	isOrphan, delay, err = dag.ProcessBlock(parentBlock, BFNone)
	if err != nil {
		t.Fatalf("TestProcessOrphans: parent block unexpectedly returned an error: %s", err)
	}
	if delay != 0 {
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
