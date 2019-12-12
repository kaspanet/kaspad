// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"fmt"
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/txscript"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/pkg/errors"
)

// CheckpointConfirmations is the number of confirmations that a good
// checkpoint candidate must be.
const CheckpointConfirmations = 2016

// newHashFromStr converts the passed big-endian hex string into a
// daghash.Hash. It only differs from the one available in daghash in that
// it ignores the error since it will only (and must only) be called with
// hard-coded, and therefore known good, hashes.
func newHashFromStr(hexStr string) *daghash.Hash {
	hash, _ := daghash.NewHashFromStr(hexStr)
	return hash
}

// newTxIDFromStr converts the passed big-endian hex string into a
// daghash.TxID. It only differs from the one available in daghash in that
// it ignores the error since it will only (and must only) be called with
// hard-coded, and therefore known good, IDs.
func newTxIDFromStr(hexStr string) *daghash.TxID {
	txID, _ := daghash.NewTxIDFromStr(hexStr)
	return txID
}

// Checkpoints returns a slice of checkpoints (regardless of whether they are
// already known). When there are no checkpoints for the DAG, it will return
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
// is already known). When there are no defined checkpoints for the active DAG
// instance, it will return nil.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) LatestCheckpoint() *dagconfig.Checkpoint {
	if !dag.HasCheckpoints() {
		return nil
	}
	return &dag.checkpoints[len(dag.checkpoints)-1]
}

// verifyCheckpoint returns whether the passed block chainHeight and hash combination
// match the checkpoint data. It also returns true if there is no checkpoint
// data for the passed block chainHeight.
func (dag *BlockDAG) verifyCheckpoint(chainHeight uint64, hash *daghash.Hash) bool {
	if !dag.HasCheckpoints() {
		return true
	}

	// Nothing to check if there is no checkpoint data for the block chainHeight.
	checkpoint, exists := dag.checkpointsByChainHeight[chainHeight]
	if !exists {
		return true
	}

	if !checkpoint.Hash.IsEqual(hash) {
		return false
	}

	log.Infof("Verified checkpoint at chainHeight %d/block %s", checkpoint.ChainHeight,
		checkpoint.Hash)
	return true
}

// findPreviousCheckpoint finds the most recent checkpoint that is already
// available in the downloaded portion of the block DAG and returns the
// associated block node. It returns nil if a checkpoint can't be found (this
// should really only happen for blocks before the first checkpoint).
//
// This function MUST be called with the DAG lock held (for reads).
func (dag *BlockDAG) findPreviousCheckpoint() (*blockNode, error) {
	if !dag.HasCheckpoints() {
		return nil, nil
	}

	// Perform the initial search to find and cache the latest known
	// checkpoint if we haven't already previously searched.
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

			// Checkpoint found. Cache it for future lookups and
			// set the next expected checkpoint accordingly.
			dag.checkpointNode = node
			if i < numCheckpoints-1 {
				dag.nextCheckpoint = &checkpoints[i+1]
			}
			return dag.checkpointNode, nil
		}

		// No known latest checkpoint. This will only happen on blocks
		// before the first known checkpoint. So, set the next expected
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

	// When there is a next checkpoint and the chainHeight of the current
	// selected tip of the DAG does not exceed it, the current checkpoint
	// lockin is still the latest known checkpoint.
	if dag.selectedTip().chainHeight < dag.nextCheckpoint.ChainHeight {
		return dag.checkpointNode, nil
	}

	// Cache the latest known checkpoint for future lookups. Note that if
	// this lookup fails something is very wrong since the DAG has already
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
		scriptClass := txscript.GetScriptClass(txOut.ScriptPubKey)
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
//  - The block must be in the DAG
//  - The block must be at least 'CheckpointConfirmations' blocks prior to the
//    selected tip
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

	// Ensure the chainHeight of the passed block and the entry for the block
	// in the DAG match. This should always be the case unless the
	// caller provided an invalid block.
	if node.chainHeight != block.ChainHeight() {
		return false, errors.Errorf("passed block chainHeight of %d does not "+
			"match the its height in the DAG: %d", block.ChainHeight(),
			node.chainHeight)
	}

	// A checkpoint must be at least CheckpointConfirmations blocks
	// before the selected tip of the DAG.
	dagChainHeight := dag.selectedTip().chainHeight
	if node.chainHeight > (dagChainHeight - CheckpointConfirmations) {
		return false, nil
	}

	// A checkpoint must be have at least one block after it.
	//
	// This should always succeed since the check above already made sure it
	// is CheckpointConfirmations back, but be safe in case the constant
	// changes.
	if len(node.children) == 0 {
		return false, nil
	}

	// A checkpoint must be have at least one block before it.
	if &node.selectedParent == nil {
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
