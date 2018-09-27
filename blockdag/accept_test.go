package blockdag

import (
	"github.com/daglabs/btcd/dagconfig"
	"testing"
)

func TestMaybeAcceptBlockErrors(t *testing.T) {
	// Create a new database and chain instance to run tests against.
	dag, teardownFunc, err := DAGSetup("haveblock",
		&dagconfig.MainNetParams)
	if err != nil {
		t.Errorf("Failed to setup DAG instance: %v", err)
		return
	}
	defer teardownFunc()

	// Test rejecting the block if its parents are missing
	orphanBlockFile := "blk_3B.dat"
	loadedBlocks, err := loadBlocks(orphanBlockFile)
	if err != nil {
		t.Fatalf("Error loading file: %s\n", orphanBlockFile)
	}
	block := loadedBlocks[0]

	err = dag.maybeAcceptBlock(block, BFNone)
	if err == nil {
		t.Errorf("Expected error but got nil")
	}
	if ruleErr, ok := err.(RuleError); !ok || ruleErr.ErrorCode != ErrPreviousBlockUnknown {
		t.Errorf("Unexpected error. Want: %s, got: %s", ErrPreviousBlockUnknown, err)
	}

	// Test rejecting the block if its parents are invalid
	blocksFile := "blk_0_to_4.dat"
	blocks, err := loadBlocks(blocksFile)
	if err != nil {
		t.Errorf("Error loading file: %s\n", err)
		return
	}

	// Add a valid block and mark it as invalid
	block1 := blocks[1]
	_, err = dag.ProcessBlock(block1, BFNone)
	if err != nil {
		t.Fatalf("Valid block unexpectedly returned an error: %s", err)
	}
	blockNode1 := dag.index.LookupNode(block1.Hash())
	dag.index.SetStatusFlags(blockNode1, statusValidateFailed)

	block2 := blocks[2]
	err = dag.maybeAcceptBlock(block2, BFNone)
	if err == nil {
		t.Errorf("Expected error but got nil")
	}
	if ruleErr, ok := err.(RuleError); !ok || ruleErr.ErrorCode != ErrInvalidAncestorBlock {
		t.Errorf("Unexpected error. Want: %s, got: %s", ErrInvalidAncestorBlock, err)
	}
}
