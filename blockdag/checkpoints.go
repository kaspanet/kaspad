// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"fmt"
	"time"

	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/util"
)

// CheckpointConfirmations is the number of blocks before the end of the current
// best block chain that a good checkpoint candidate must be.
const CheckpointConfirmations = 2016

// newHashFromStr converts the passed big-endian hex string into a
// daghash.Hash.  It only differs from the one available in daghash in that
// it ignores the error since it will only (and must only) be called with
// hard-coded, and therefore known good, hashes.
func newHashFromStr(hexStr string) *daghash.Hash {
	hash, _ := daghash.NewHashFromStr(hexStr)
	return hash
}

// newTxIDFromStr converts the passed big-endian hex string into a
// daghash.TxID.  It only differs from the one available in daghash in that
// it ignores the error since it will only (and must only) be called with
// hard-coded, and therefore known good, IDs.
func newTxIDFromStr(hexStr string) *daghash.TxID {
	txID, _ := daghash.NewTxIDFromStr(hexStr)
	return txID
}

// Checkpoints returns a slice of checkpoints (regardless of whether they are
// already known).  When there are no checkpoints for the chain, it will return
// nil.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) Checkpoints() []dagconfig.Checkpoint {
	return dag.checkpoints
}

// HasCheckpoints returns whether this BlockDAG has checkpoints defined.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) HasCheckpoints() bool {
	return len(dag.checkpoints) > 0
}

// LatestCheckpoint returns the most recent checkpoint (regardless of whether it
// is already known). When there are no defined checkpoints for the active chain
// instance, it will return nil.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) LatestCheckpoint() *dagconfig.Checkpoint {
	if !dag.HasCheckpoints() {
		return nil
	}
	return &dag.checkpoints[len(dag.checkpoints)-1]
}

// verifyCheckpoint returns whether the passed block height and hash combination
// match the checkpoint data.  It also returns true if there is no checkpoint
// data for the passed block height.
func (dag *BlockDAG) verifyCheckpoint(height uint64, hash *daghash.Hash) bool {
	if !dag.HasCheckpoints() {
		return true
	}

	// Nothing to check if there is no checkpoint data for the block height.
	checkpoint, exists := dag.checkpointsByHeight[height]
	if !exists {
		return true
	}

	if !checkpoint.Hash.IsEqual(hash) {
		return false
	}

	log.Infof("Verified checkpoint at height %d/block %s", checkpoint.Height,
		checkpoint.Hash)
	return true
}

// findPreviousCheckpoint finds the most recent checkpoint that is already
// available in the downloaded portion of the block chain and returns the
// associated block node.  It returns nil if a checkpoint can't be found (this
// should really only happen for blocks before the first checkpoint).
//
// This function MUST be called with the DAG lock held (for reads).
func (dag *BlockDAG) findPreviousCheckpoint() (*blockNode, error) {
	if !dag.HasCheckpoints() {
		return nil, nil
	}

	// Perform the initial search to find and cache the latest known
	// checkpoint if the best chain is not known yet or we haven't already
	// previously searched.
	checkpoints := dag.checkpoints
	numCheckpoints := len(checkpoints)
	if dag.checkpointNode == nil && dag.nextCheckpoint == nil {
		// Loop backwards through the available checkpoints to find one
		// that is already available.
		for i := numCheckpoints - 1; i >= 0; i-- {
			node := dag.index.LookupNode(checkpoints[i].Hash)
			if node == nil {
				continue
			}

			// Checkpoint found.  Cache it for future lookups and
			// set the next expected checkpoint accordingly.
			dag.checkpointNode = node
			if i < numCheckpoints-1 {
				dag.nextCheckpoint = &checkpoints[i+1]
			}
			return dag.checkpointNode, nil
		}

		// No known latest checkpoint.  This will only happen on blocks
		// before the first known checkpoint.  So, set the next expected
		// checkpoint to the first checkpoint and return the fact there
		// is no latest known checkpoint block.
		dag.nextCheckpoint = &checkpoints[0]
		return nil, nil
	}

	// At this point we've already searched for the latest known checkpoint,
	// so when there is no next checkpoint, the current checkpoint lockin
	// will always be the latest known checkpoint.
	if dag.nextCheckpoint == nil {
		return dag.checkpointNode, nil
	}

	// When there is a next checkpoint and the height of the current best
	// chain does not exceed it, the current checkpoint lockin is still
	// the latest known checkpoint.
	if dag.selectedTip().height < dag.nextCheckpoint.Height {
		return dag.checkpointNode, nil
	}

	// We've reached or exceeded the next checkpoint height.  Note that
	// once a checkpoint lockin has been reached, forks are prevented from
	// any blocks before the checkpoint, so we don't have to worry about the
	// checkpoint going away out from under us due to a chain reorganize.

	// Cache the latest known checkpoint for future lookups.  Note that if
	// this lookup fails something is very wrong since the chain has already
	// passed the checkpoint which was verified as accurate before inserting
	// it.
	checkpointNode := dag.index.LookupNode(dag.nextCheckpoint.Hash)
	if checkpointNode == nil {
		return nil, AssertError(fmt.Sprintf("findPreviousCheckpoint "+
			"failed lookup of known good block node %s",
			dag.nextCheckpoint.Hash))
	}
	dag.checkpointNode = checkpointNode

	// Set the next expected checkpoint.
	checkpointIndex := -1
	for i := numCheckpoints - 1; i >= 0; i-- {
		if checkpoints[i].Hash.IsEqual(dag.nextCheckpoint.Hash) {
			checkpointIndex = i
			break
		}
	}
	dag.nextCheckpoint = nil
	if checkpointIndex != -1 && checkpointIndex < numCheckpoints-1 {
		dag.nextCheckpoint = &checkpoints[checkpointIndex+1]
	}

	return dag.checkpointNode, nil
}

// isNonstandardTransaction determines whether a transaction contains any
// scripts which are not one of the standard types.
func isNonstandardTransaction(tx *util.Tx) bool {
	// Check all of the output public key scripts for non-standard scripts.
	for _, txOut := range tx.MsgTx().TxOut {
		scriptClass := txscript.GetScriptClass(txOut.PkScript)
		if scriptClass == txscript.NonStandardTy {
			return true
		}
	}
	return false
}

// IsCheckpointCandidate returns whether or not the passed block is a good
// checkpoint candidate.
//
// The factors used to determine a good checkpoint are:
//  - The block must be in the main chain
//  - The block must be at least 'CheckpointConfirmations' blocks prior to the
//    current end of the main chain
//  - The timestamps for the blocks before and after the checkpoint must have
//    timestamps which are also before and after the checkpoint, respectively
//    (due to the median time allowance this is not always the case)
//  - The block must not contain any strange transaction such as those with
//    nonstandard scripts
//
// The intent is that candidates are reviewed by a developer to make the final
// decision and then manually added to the list of checkpoints for a network.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) IsCheckpointCandidate(block *util.Block) (bool, error) {
	dag.dagLock.RLock()
	defer dag.dagLock.RUnlock()

	// A checkpoint must be in the DAG.
	node := dag.index.LookupNode(block.Hash())
	if node == nil {
		return false, nil
	}

	// Ensure the height of the passed block and the entry for the block in
	// the main chain match.  This should always be the case unless the
	// caller provided an invalid block.
	if node.height != block.Height() {
		return false, fmt.Errorf("passed block height of %d does not "+
			"match the main chain height of %d", block.Height(),
			node.height)
	}

	// A checkpoint must be at least CheckpointConfirmations blocks
	// before the end of the main chain.
	dagHeight := dag.selectedTip().height
	if node.height > (dagHeight - CheckpointConfirmations) {
		return false, nil
	}

	// A checkpoint must be have at least one block after it.
	//
	// This should always succeed since the check above already made sure it
	// is CheckpointConfirmations back, but be safe in case the constant
	// changes.
	nextNode := node.diffChild
	if nextNode == nil {
		return false, nil
	}

	// A checkpoint must be have at least one block before it.
	if &node.selectedParent == nil {
		return false, nil
	}

	// A checkpoint must have timestamps for the block and the blocks on
	// either side of it in order (due to the median time allowance this is
	// not always the case).
	prevTime := time.Unix(node.selectedParent.timestamp, 0)
	curTime := block.MsgBlock().Header.Timestamp
	nextTime := time.Unix(nextNode.timestamp, 0)
	if prevTime.After(curTime) || nextTime.Before(curTime) {
		return false, nil
	}

	// A checkpoint must have transactions that only contain standard
	// scripts.
	for _, tx := range block.Transactions() {
		if isNonstandardTransaction(tx) {
			return false, nil
		}
	}

	// All of the checks passed, so the block is a candidate.
	return true, nil
}
