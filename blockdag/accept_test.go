package blockdag

import (
	"errors"
	"strings"
	"testing"

	"bou.ke/monkey"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/util"
)

func TestMaybeAcceptBlockErrors(t *testing.T) {
	// Create a new database and DAG instance to run tests against.
	dag, teardownFunc, err := DAGSetup("TestMaybeAcceptBlockErrors", Config{
		DAGParams: &dagconfig.SimNetParams,
	})
	if err != nil {
		t.Fatalf("TestMaybeAcceptBlockErrors: Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()

	dag.TstSetCoinbaseMaturity(1)

	// Test rejecting the block if its parents are missing
	orphanBlockFile := "blk_3B.dat"
	loadedBlocks, err := loadBlocks(orphanBlockFile)
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
	ruleErr, ok := err.(RuleError)
	if !ok {
		t.Errorf("TestMaybeAcceptBlockErrors: rejecting the block if its parents are missing: "+
			"Expected RuleError but got %s", err)
	} else if ruleErr.ErrorCode != ErrParentBlockUnknown {
		t.Errorf("TestMaybeAcceptBlockErrors: rejecting the block if its parents are missing: "+
			"Unexpected error code. Want: %s, got: %s", ErrParentBlockUnknown, ruleErr.ErrorCode)
	}

	// Test rejecting the block if its parents are invalid
	blocksFile := "blk_0_to_4.dat"
	blocks, err := loadBlocks(blocksFile)
	if err != nil {
		t.Fatalf("TestMaybeAcceptBlockErrors: "+
			"Error loading file '%s': %s\n", blocksFile, err)
	}

	// Add a valid block and mark it as invalid
	block1 := blocks[1]
	isOrphan, err := dag.ProcessBlock(block1, BFNone)
	if err != nil {
		t.Fatalf("TestMaybeAcceptBlockErrors: Valid block unexpectedly returned an error: %s", err)
	}
	if isOrphan {
		t.Fatalf("TestMaybeAcceptBlockErrors: incorrectly returned block 1 is an orphan")
	}
	blockNode1 := dag.index.LookupNode(block1.Hash())
	dag.index.SetStatusFlags(blockNode1, statusValidateFailed)

	block2 := blocks[2]
	err = dag.maybeAcceptBlock(block2, BFNone)
	if err == nil {
		t.Errorf("TestMaybeAcceptBlockErrors: rejecting the block if its parents are invalid: "+
			"Expected: %s, got: <nil>", ErrInvalidAncestorBlock)
	}
	ruleErr, ok = err.(RuleError)
	if !ok {
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
	ruleErr, ok = err.(RuleError)
	if !ok {
		t.Errorf("TestMaybeAcceptBlockErrors: rejecting the block due to bad context: "+
			"Expected RuleError but got %s", err)
	} else if ruleErr.ErrorCode != ErrUnexpectedDifficulty {
		t.Errorf("TestMaybeAcceptBlockErrors: rejecting the block due to bad context: "+
			"Unexpected error. Want: %s, got: %s", ErrUnexpectedDifficulty, ruleErr.ErrorCode)
	}

	// Set block2's bits back to valid for next tests
	block2.MsgBlock().Header.Bits = originalBits

	// Test rejecting the node due to database error
	databaseErrorMessage := "database error"
	guard := monkey.Patch(dbStoreBlock, func(dbTx database.Tx, block *util.Block) error {
		return errors.New(databaseErrorMessage)
	})
	defer guard.Unpatch()
	err = dag.maybeAcceptBlock(block2, BFNone)
	if err == nil {
		t.Errorf("TestMaybeAcceptBlockErrors: rejecting the node due to database error: "+
			"Expected: %s, got: <nil>", databaseErrorMessage)
	}
	if !strings.Contains(err.Error(), databaseErrorMessage) {
		t.Errorf("TestMaybeAcceptBlockErrors: rejecting the node due to database error: "+
			"Unexpected error. Want: %s, got: %s", databaseErrorMessage, err)
	}
	guard.Unpatch()

	// Test rejecting the node due to index error
	indexErrorMessage := "index error"
	guard = monkey.Patch((*blockIndex).flushToDB, func(_ *blockIndex) error {
		return errors.New(indexErrorMessage)
	})
	defer guard.Unpatch()
	err = dag.maybeAcceptBlock(block2, BFNone)
	if err == nil {
		t.Errorf("TestMaybeAcceptBlockErrors: rejecting the node due to index error: "+
			"Expected %s, got: <nil>", indexErrorMessage)
	}
	if !strings.Contains(err.Error(), indexErrorMessage) {
		t.Errorf("TestMaybeAcceptBlockErrors: rejecting the node due to index error: "+
			"Unexpected error. Want: %s, got: %s", indexErrorMessage, err)
	}
}
