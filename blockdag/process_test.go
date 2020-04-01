package blockdag

import (
	"github.com/kaspanet/kaspad/util"
	"path/filepath"
	"testing"
	"time"

	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/util/daghash"
)

func TestProcessOrphans(t *testing.T) {
	dag, teardownFunc, err := DAGSetup("TestProcessOrphans", true, Config{
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

func TestProcessDelayedBlocks(t *testing.T) {
	// We use dag1 so we can build the test blocks with the proper
	// block header (UTXO commitment, acceptedIDMerkleroot, etc), and
	// then we use dag2 for the actual test.
	dag1, teardownFunc, err := DAGSetup("TestProcessDelayedBlocks1", true, Config{
		DAGParams: &dagconfig.SimnetParams,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	isDAG1Open := true
	defer func() {
		if isDAG1Open {
			teardownFunc()
		}
	}()

	initialTime := dag1.dagParams.GenesisBlock.Header.Timestamp
	// Here we use a fake time source that returns a timestamp
	// one hour into the future to make delayedBlock artificially
	// valid.
	dag1.timeSource = newFakeTimeSource(initialTime.Add(time.Hour))

	delayedBlock, err := PrepareBlockForTest(dag1, []*daghash.Hash{dag1.dagParams.GenesisBlock.BlockHash()}, nil)
	if err != nil {
		t.Fatalf("error in PrepareBlockForTest: %s", err)
	}

	blockDelay := time.Duration(dag1.dagParams.TimestampDeviationTolerance*uint64(dag1.targetTimePerBlock)+5) * time.Second
	delayedBlock.Header.Timestamp = initialTime.Add(blockDelay)

	isOrphan, isDelayed, err := dag1.ProcessBlock(util.NewBlock(delayedBlock), BFNoPoWCheck)
	if err != nil {
		t.Fatalf("ProcessBlock returned unexpected error: %s\n", err)
	}
	if isOrphan {
		t.Fatalf("ProcessBlock incorrectly returned delayedBlock " +
			"is an orphan\n")
	}
	if isDelayed {
		t.Fatalf("ProcessBlock incorrectly returned delayedBlock " +
			"is delayed\n")
	}

	delayedBlockChild, err := PrepareBlockForTest(dag1, []*daghash.Hash{delayedBlock.BlockHash()}, nil)
	if err != nil {
		t.Fatalf("error in PrepareBlockForTest: %s", err)
	}

	teardownFunc()
	isDAG1Open = false

	// Here the actual test begins. We add a delayed block and
	// its child and check that they are not added to the DAG,
	// and check that they're added only if we add a new block
	// after the delayed block timestamp is valid.
	dag2, teardownFunc2, err := DAGSetup("TestProcessDelayedBlocks2", true, Config{
		DAGParams: &dagconfig.SimnetParams,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc2()
	dag2.timeSource = newFakeTimeSource(initialTime)

	isOrphan, isDelayed, err = dag2.ProcessBlock(util.NewBlock(delayedBlock), BFNoPoWCheck)
	if err != nil {
		t.Fatalf("ProcessBlock returned unexpected error: %s\n", err)
	}
	if isOrphan {
		t.Fatalf("ProcessBlock incorrectly returned delayedBlock " +
			"is an orphan\n")
	}
	if !isDelayed {
		t.Fatalf("ProcessBlock incorrectly returned delayedBlock " +
			"is not delayed\n")
	}

	if dag2.IsInDAG(delayedBlock.BlockHash()) {
		t.Errorf("dag.IsInDAG should return false for a delayed block")
	}
	if !dag2.IsKnownBlock(delayedBlock.BlockHash()) {
		t.Errorf("dag.IsKnownBlock should return true for a a delayed block")
	}

	isOrphan, isDelayed, err = dag2.ProcessBlock(util.NewBlock(delayedBlockChild), BFNoPoWCheck)
	if err != nil {
		t.Fatalf("ProcessBlock returned unexpected error: %s\n", err)
	}
	if isOrphan {
		t.Fatalf("ProcessBlock incorrectly returned delayedBlockChild " +
			"is an orphan\n")
	}
	if !isDelayed {
		t.Fatalf("ProcessBlock incorrectly returned delayedBlockChild " +
			"is not delayed\n")
	}

	if dag2.IsInDAG(delayedBlockChild.BlockHash()) {
		t.Errorf("dag.IsInDAG should return false for a child of a delayed block")
	}
	if !dag2.IsKnownBlock(delayedBlockChild.BlockHash()) {
		t.Errorf("dag.IsKnownBlock should return true for a child of a delayed block")
	}

	blockBeforeDelay, err := PrepareBlockForTest(dag2, []*daghash.Hash{dag2.dagParams.GenesisBlock.BlockHash()}, nil)
	if err != nil {
		t.Fatalf("error in PrepareBlockForTest: %s", err)
	}
	isOrphan, isDelayed, err = dag2.ProcessBlock(util.NewBlock(blockBeforeDelay), BFNoPoWCheck)
	if err != nil {
		t.Fatalf("ProcessBlock returned unexpected error: %s\n", err)
	}
	if isOrphan {
		t.Fatalf("ProcessBlock incorrectly returned blockBeforeDelay " +
			"is an orphan\n")
	}
	if isDelayed {
		t.Fatalf("ProcessBlock incorrectly returned blockBeforeDelay " +
			"is delayed\n")
	}

	if dag2.IsInDAG(delayedBlock.BlockHash()) {
		t.Errorf("delayedBlock shouldn't be added to the DAG because its time hasn't reached yet")
	}
	if dag2.IsInDAG(delayedBlockChild.BlockHash()) {
		t.Errorf("delayedBlockChild shouldn't be added to the DAG because its parent is not in the DAG")
	}

	// We advance the clock to the point where delayedBlock timestamp is valid.
	deviationTolerance := int64(dag2.TimestampDeviationTolerance) * dag2.targetTimePerBlock
	secondsUntilDelayedBlockIsValid := delayedBlock.Header.Timestamp.Unix() - deviationTolerance - dag2.Now().Unix() + 1
	dag2.timeSource = newFakeTimeSource(initialTime.Add(time.Duration(secondsUntilDelayedBlockIsValid) * time.Second))

	blockAfterDelay, err := PrepareBlockForTest(dag2,
		[]*daghash.Hash{dag2.dagParams.GenesisBlock.BlockHash()},
		nil)
	if err != nil {
		t.Fatalf("error in PrepareBlockForTest: %s", err)
	}
	isOrphan, isDelayed, err = dag2.ProcessBlock(util.NewBlock(blockAfterDelay), BFNoPoWCheck)
	if err != nil {
		t.Fatalf("ProcessBlock returned unexpected error: %s\n", err)
	}
	if isOrphan {
		t.Fatalf("ProcessBlock incorrectly returned blockBeforeDelay " +
			"is an orphan\n")
	}
	if isDelayed {
		t.Fatalf("ProcessBlock incorrectly returned blockBeforeDelay " +
			"is not delayed\n")
	}

	if !dag2.IsInDAG(delayedBlock.BlockHash()) {
		t.Fatalf("delayedBlock should be added to the DAG because its time has been reached")
	}
	if !dag2.IsInDAG(delayedBlockChild.BlockHash()) {
		t.Errorf("delayedBlockChild shouldn't be added to the DAG because its parent has been added to the DAG")
	}
}
