// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"fmt"
	"math"

	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/wire"
)

func validateParents(blockHeader *wire.BlockHeader, parents blockSet) error {
	minHeight := int32(math.MaxInt32)
	queue := NewHeap()
	visited := newSet()
	for _, parent := range parents {
		if parent.height < minHeight {
			minHeight = parent.height
		}
		for _, grandParent := range parent.parents {
			if !visited.contains(grandParent) {
				queue.Push(grandParent)
				visited.add(grandParent)
			}
		}
	}
	for queue.Len() > 0 {
		current := queue.Pop()
		if parents.contains(current) {
			return fmt.Errorf("Block %s is both a parent of %s and an ancestor of another parent", current.hash, blockHeader.BlockHash())
		}
		if current.height > minHeight {
			for _, parent := range current.parents {
				if !visited.contains(parent) {
					queue.Push(current)
					visited.add(current)
				}
			}
		}
	}
	return nil
}

// maybeAcceptBlock potentially accepts a block into the block DAG. It
// performs several validation checks which depend on its position within
// the block DAG before adding it. The block is expected to have already
// gone through ProcessBlock before calling this function with it.
//
// The flags are also passed to checkBlockContext and connectToDAG.  See
// their documentation for how the flags modify their behavior.
//
// This function MUST be called with the dagLock held (for writes).
func (dag *BlockDAG) maybeAcceptBlock(block *util.Block, flags BehaviorFlags) error {
	// The height of this block is one more than the referenced previous
	// block.
	parents, err := lookupPreviousNodes(block, dag)
	if err != nil {
		return err
	}

	selectedParent := parents.first() //TODO (Ori): This is wrong, done only for compilation
	blockHeight := parents.maxHeight() + 1
	block.SetHeight(blockHeight)

	// The block must pass all of the validation rules which depend on the
	// position of the block within the block DAG.
	err = dag.checkBlockContext(block, selectedParent, flags)
	if err != nil {
		return err
	}

	// Insert the block into the database if it's not already there.  Even
	// though it is possible the block will ultimately fail to connect, it
	// has already passed all proof-of-work and validity tests which means
	// it would be prohibitively expensive for an attacker to fill up the
	// disk with a bunch of blocks that fail to connect.  This is necessary
	// since it allows block download to be decoupled from the much more
	// expensive connection logic.  It also has some other nice properties
	// such as making blocks that never become part of the main chain or
	// blocks that fail to connect available for further analysis.
	err = dag.db.Update(func(dbTx database.Tx) error {
		return dbStoreBlock(dbTx, block)
	})
	if err != nil {
		return err
	}

	// Create a new block node for the block and add it to the node index. Even
	// if the block ultimately gets connected to the main chain, it starts out
	// on a side chain.
	blockHeader := &block.MsgBlock().Header
	err = validateParents(blockHeader, parents)
	if err != nil {
		return err
	}
	newNode := newBlockNode(blockHeader, parents, dag.dagParams.K)
	newNode.status = statusDataStored

	dag.index.AddNode(newNode)
	err = dag.index.flushToDB()
	if err != nil {
		return err
	}

	// Connect the passed block to the DAG. This also handles validation of the
	// transaction scripts.
	err = dag.connectToDAG(newNode, parents, block, flags)
	if err != nil {
		return err
	}

	// Notify the caller that the new block was accepted into the block
	// chain.  The caller would typically want to react by relaying the
	// inventory to other peers.
	dag.dagLock.Unlock()
	dag.sendNotification(NTBlockAccepted, block)
	dag.dagLock.Lock()

	return nil
}

func lookupPreviousNodes(block *util.Block, blockDAG *BlockDAG) (blockSet, error) {
	header := block.MsgBlock().Header
	prevHashes := header.PrevBlocks

	nodes := newSet()
	for _, prevHash := range prevHashes {
		node := blockDAG.index.LookupNode(&prevHash)
		if node == nil {
			str := fmt.Sprintf("previous block %s is unknown", prevHashes)
			return nil, ruleError(ErrPreviousBlockUnknown, str)
		} else if blockDAG.index.NodeStatus(node).KnownInvalid() {
			str := fmt.Sprintf("previous block %s is known to be invalid", prevHashes)
			return nil, ruleError(ErrInvalidAncestorBlock, str)
		}

		nodes.add(node)
	}

	return nodes, nil
}
