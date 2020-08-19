package blockdag

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/util"

	"github.com/kaspanet/kaspad/domain/dagconfig"
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
	node, ok := dag.index.LookupNode(childBlock.Hash())
	if !ok {
		t.Fatalf("TestProcessOrphans: child block missing from block index")
	}
	if !dag.index.BlockNodeStatus(node).KnownInvalid() {
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

	initialTime := dag1.Params.GenesisBlock.Header.Timestamp
	// Here we use a fake time source that returns a timestamp
	// one hour into the future to make delayedBlock artificially
	// valid.
	dag1.timeSource = newFakeTimeSource(initialTime.Add(time.Hour))

	delayedBlock, err := PrepareBlockForTest(dag1, []*daghash.Hash{dag1.Params.GenesisBlock.BlockHash()}, nil)
	if err != nil {
		t.Fatalf("error in PrepareBlockForTest: %s", err)
	}

	blockDelay := time.Duration(dag1.Params.TimestampDeviationTolerance)*dag1.Params.TargetTimePerBlock + 5*time.Second
	delayedBlock.Header.Timestamp = initialTime.Add(blockDelay)

	// We change the nonce here because processDelayedBlocks always runs without BFNoPoWCheck.
	delayedBlock.Header.Nonce = 2

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

	blockBeforeDelay, err := PrepareBlockForTest(dag2, []*daghash.Hash{dag2.Params.GenesisBlock.BlockHash()}, nil)
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
	deviationTolerance := time.Duration(dag2.TimestampDeviationTolerance) * dag2.Params.TargetTimePerBlock
	timeUntilDelayedBlockIsValid := delayedBlock.Header.Timestamp.
		Add(-deviationTolerance).
		Sub(dag2.Now()) +
		time.Second
	dag2.timeSource = newFakeTimeSource(initialTime.Add(timeUntilDelayedBlockIsValid))

	blockAfterDelay, err := PrepareBlockForTest(dag2,
		[]*daghash.Hash{dag2.Params.GenesisBlock.BlockHash()},
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

func TestMaybeAcceptBlockErrors(t *testing.T) {
	// Create a new database and DAG instance to run tests against.
	dag, teardownFunc, err := DAGSetup("TestMaybeAcceptBlockErrors", true, Config{
		DAGParams: &dagconfig.SimnetParams,
	})
	if err != nil {
		t.Fatalf("TestMaybeAcceptBlockErrors: Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()

	dag.TestSetCoinbaseMaturity(0)

	// Test rejecting the block if its parents are missing
	orphanBlockFile := "blk_3B.dat"
	loadedBlocks, err := LoadBlocks(filepath.Join("testdata/", orphanBlockFile))
	if err != nil {
		t.Fatalf("TestMaybeAcceptBlockErrors: "+
			"Error loading file '%s': %s\n", orphanBlockFile, err)
	}
	block := loadedBlocks[0]

	err = dag.maybeAcceptBlock(block, BFNone)
	if err == nil {
		t.Errorf("TestMaybeAcceptBlockErrors: rejecting the block if its parents are missing: "+
			"Expected: %s, got: <nil>", ErrParentBlockUnknown)
	}
	var ruleErr RuleError
	if ok := errors.As(err, &ruleErr); !ok {
		t.Errorf("TestMaybeAcceptBlockErrors: rejecting the block if its parents are missing: "+
			"Expected RuleError but got %s", err)
	} else if ruleErr.ErrorCode != ErrParentBlockUnknown {
		t.Errorf("TestMaybeAcceptBlockErrors: rejecting the block if its parents are missing: "+
			"Unexpected error code. Want: %s, got: %s", ErrParentBlockUnknown, ruleErr.ErrorCode)
	}

	// Test rejecting the block if its parents are invalid
	blocksFile := "blk_0_to_4.dat"
	blocks, err := LoadBlocks(filepath.Join("testdata/", blocksFile))
	if err != nil {
		t.Fatalf("TestMaybeAcceptBlockErrors: "+
			"Error loading file '%s': %s\n", blocksFile, err)
	}

	// Add a valid block and mark it as invalid
	block1 := blocks[1]
	isOrphan, isDelayed, err := dag.ProcessBlock(block1, BFNone)
	if err != nil {
		t.Fatalf("TestMaybeAcceptBlockErrors: Valid block unexpectedly returned an error: %s", err)
	}
	if isDelayed {
		t.Fatalf("TestMaybeAcceptBlockErrors: block 1 is too far in the future")
	}
	if isOrphan {
		t.Fatalf("TestMaybeAcceptBlockErrors: incorrectly returned block 1 is an orphan")
	}
	blockNode1, ok := dag.index.LookupNode(block1.Hash())
	if !ok {
		t.Fatalf("block %s does not exist in the DAG", block1.Hash())
	}
	dag.index.SetBlockNodeStatus(blockNode1, statusValidateFailed)

	block2 := blocks[2]
	err = dag.maybeAcceptBlock(block2, BFNone)
	if err == nil {
		t.Errorf("TestMaybeAcceptBlockErrors: rejecting the block if its parents are invalid: "+
			"Expected: %s, got: <nil>", ErrInvalidAncestorBlock)
	}
	if ok := errors.As(err, &ruleErr); !ok {
		t.Errorf("TestMaybeAcceptBlockErrors: rejecting the block if its parents are invalid: "+
			"Expected RuleError but got %s", err)
	} else if ruleErr.ErrorCode != ErrInvalidAncestorBlock {
		t.Errorf("TestMaybeAcceptBlockErrors: rejecting the block if its parents are invalid: "+
			"Unexpected error. Want: %s, got: %s", ErrInvalidAncestorBlock, ruleErr.ErrorCode)
	}

	// Set block1's status back to valid for next tests
	dag.index.SetBlockNodeStatus(blockNode1, statusValid)

	// Test rejecting the block due to bad context
	originalBits := block2.MsgBlock().Header.Bits
	block2.MsgBlock().Header.Bits = 0
	err = dag.maybeAcceptBlock(block2, BFNone)
	if err == nil {
		t.Errorf("TestMaybeAcceptBlockErrors: rejecting the block due to bad context: "+
			"Expected: %s, got: <nil>", ErrUnexpectedDifficulty)
	}
	if ok := errors.As(err, &ruleErr); !ok {
		t.Errorf("TestMaybeAcceptBlockErrors: rejecting the block due to bad context: "+
			"Expected RuleError but got %s", err)
	} else if ruleErr.ErrorCode != ErrUnexpectedDifficulty {
		t.Errorf("TestMaybeAcceptBlockErrors: rejecting the block due to bad context: "+
			"Unexpected error. Want: %s, got: %s", ErrUnexpectedDifficulty, ruleErr.ErrorCode)
	}

	// Set block2's bits back to valid for next tests
	block2.MsgBlock().Header.Bits = originalBits
}
