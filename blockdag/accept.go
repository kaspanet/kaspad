// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"fmt"

	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/util"
)

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
	parents, err := lookupParentNodes(block, dag)
	if err != nil {
		return err
	}

	bluestParent := parents.bluest()
	blockHeight := int32(0)
	if !block.IsGenesis() {
		blockHeight = parents.maxHeight() + 1
	}
	block.SetHeight(blockHeight)

	// The block must pass all of the validation rules which depend on the
	// position of the block within the block DAG.
	err = dag.checkBlockContext(block, parents, bluestParent, flags)
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
	// such as making blocks that never become part of the DAG or
	// blocks that fail to connect available for further analysis.
	err = dag.db.Update(func(dbTx database.Tx) error {
		return dbStoreBlock(dbTx, block)
	})
	if err != nil {
		return err
	}

	// Create a new block node for the block and add it to the node index.
	blockHeader := &block.MsgBlock().Header
	newNode := newBlockNode(blockHeader, parents, dag.dagParams.K)
	newNode.status = statusDataStored
	// newBlockNode adds node into children maps of its parents. So it must be
	// removed in case of error.
	isOk := false
	defer func() {
		if !isOk {
			newNode.detachFromParents()
		}
	}()

	dag.index.AddNode(newNode)
	err = dag.index.flushToDB()
	if err != nil {
		return err
	}

	// Connect the passed block to the DAG. This also handles validation of the
	// transaction scripts.
	err = dag.addBlock(newNode, parents, block, flags)
	if err != nil {
		return err
	}

	// Notify the caller that the new block was accepted into the block
	// DAG.  The caller would typically want to react by relaying the
	// inventory to other peers.
	dag.dagLock.Unlock()
	dag.sendNotification(NTBlockAccepted, block)
	dag.dagLock.Lock()

	isOk = true

	return nil
}

func lookupParentNodes(block *util.Block, blockDAG *BlockDAG) (blockSet, error) {
	header := block.MsgBlock().Header
	parentHashes := header.ParentHashes

	nodes := newSet()
	for _, parentHash := range parentHashes {
		node := blockDAG.index.LookupNode(&parentHash)
		if node == nil {
			str := fmt.Sprintf("parent block %s is unknown", parentHashes)
			return nil, ruleError(ErrParentBlockUnknown, str)
		} else if blockDAG.index.NodeStatus(node).KnownInvalid() {
			str := fmt.Sprintf("parent block %s is known to be invalid", parentHashes)
			return nil, ruleError(ErrInvalidAncestorBlock, str)
		}

		nodes.add(node)
	}

	return nodes, nil
}
