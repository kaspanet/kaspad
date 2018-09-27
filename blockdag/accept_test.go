package blockdag

import (
	"bou.ke/monkey"
	"errors"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/util"
	"strings"
	"testing"
)

func TestMaybeAcceptBlockErrors(t *testing.T) {
	// Create a new database and DAG instance to run tests against.
	dag, teardownFunc, err := DAGSetup("TestMaybeAcceptBlockErrors", &dagconfig.MainNetParams)
	if err != nil {
		t.Errorf("Failed to setup DAG instance: %v", err)
		return
	}
	defer teardownFunc()

	dag.TstSetCoinbaseMaturity(1)

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

	// Set block1's status back to valid for next tests
	dag.index.UnsetStatusFlags(blockNode1, statusValidateFailed)

	// Test rejecting the block due to bad context
	originalBits := block2.MsgBlock().Header.Bits
	block2.MsgBlock().Header.Bits = 0
	err = dag.maybeAcceptBlock(block2, BFNone)
	if err == nil {
		t.Errorf("Expected error but got nil")
	}
	if ruleErr, ok := err.(RuleError); !ok || ruleErr.ErrorCode != ErrUnexpectedDifficulty {
		t.Errorf("Unexpected error. Want: %s, got: %s", ErrUnexpectedDifficulty, err)
	}

	// Set block2's bits back to valid for next tests
	block2.MsgBlock().Header.Bits = originalBits

	// Test rejecting the node due to database error
	databaseErrorMessage := "database error"
	monkey.Patch(dbStoreBlock, func(dbTx database.Tx, block *util.Block) error {
		return errors.New(databaseErrorMessage)
	})
	err = dag.maybeAcceptBlock(block2, BFNone)
	if err == nil {
		t.Errorf("Expected error but got nil")
	}
	if !strings.Contains(err.Error(), databaseErrorMessage) {
		t.Errorf("Unexpected error. Want: %s, got: %s", databaseErrorMessage, err)
	}
	monkey.Unpatch(dbStoreBlock)

	// Test rejecting the node due to index error
	indexErrorMessage := "index error"
	monkey.Patch((*blockIndex).flushToDB, func(_ *blockIndex) error {
		return errors.New(indexErrorMessage)
	})
	err = dag.maybeAcceptBlock(block2, BFNone)
	if err == nil {
		t.Errorf("Expected error but got nil")
	}
	if !strings.Contains(err.Error(), indexErrorMessage) {
		t.Errorf("Unexpected error. Want: %s, got: %s", indexErrorMessage, err)
	}
	monkey.Unpatch((*blockIndex).flushToDB)
}
