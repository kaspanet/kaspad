package blockdag

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/kaspanet/kaspad/dagconfig"
)

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
	dag.index.SetStatusFlags(blockNode1, statusValidateFailed)

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
	dag.index.UnsetStatusFlags(blockNode1, statusValidateFailed)

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
