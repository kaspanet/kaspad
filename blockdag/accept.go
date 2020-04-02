// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"fmt"
	"github.com/kaspanet/kaspad/dbaccess"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
)

func (dag *BlockDAG) addNodeToIndexWithInvalidAncestor(block *util.Block) error {
	blockHeader := &block.MsgBlock().Header
	newNode, _ := dag.newBlockNode(blockHeader, newBlockSet())
	newNode.status = statusInvalidAncestor
	dag.index.AddNode(newNode)

	dbTx, err := dbaccess.NewTx()
	if err != nil {
		return err
	}
	defer dbTx.RollbackUnlessClosed()
	err = dag.index.flushToDB(dbTx)
	if err != nil {
		return err
	}
	return dbTx.Commit()
}

// maybeAcceptBlock potentially accepts a block into the block DAG. It
// performs several validation checks which depend on its position within
// the block DAG before adding it. The block is expected to have already
// gone through ProcessBlock before calling this function with it.
//
// The flags are also passed to checkBlockContext and connectToDAG. See
// their documentation for how the flags modify their behavior.
//
// This function MUST be called with the dagLock held (for writes).
func (dag *BlockDAG) maybeAcceptBlock(block *util.Block, flags BehaviorFlags) error {
	parents, err := lookupParentNodes(block, dag)
	if err != nil {
		var ruleErr RuleError
		if ok := errors.As(err, &ruleErr); ok && ruleErr.ErrorCode == ErrInvalidAncestorBlock {
			err := dag.addNodeToIndexWithInvalidAncestor(block)
			if err != nil {
				return err
			}
		}
		return err
	}

	// The block must pass all of the validation rules which depend on the
	// position of the block within the block DAG.
	err = dag.checkBlockContext(block, parents, flags)
	if err != nil {
		return err
	}

	// Create a new block node for the block and add it to the node index.
	newNode, selectedParentAnticone := dag.newBlockNode(&block.MsgBlock().Header, parents)
	newNode.status = statusDataStored
	dag.index.AddNode(newNode)

	// Insert the block into the database if it's not already there. Even
	// though it is possible the block will ultimately fail to connect, it
	// has already passed all proof-of-work and validity tests which means
	// it would be prohibitively expensive for an attacker to fill up the
	// disk with a bunch of blocks that fail to connect. This is necessary
	// since it allows block download to be decoupled from the much more
	// expensive connection logic. It also has some other nice properties
	// such as making blocks that never become part of the DAG or
	// blocks that fail to connect available for further analysis.
	dbTx, err := dbaccess.NewTx()
	if err != nil {
		return err
	}
	defer dbTx.RollbackUnlessClosed()
	blockExists, err := dbaccess.HasBlock(dbTx, block.Hash())
	if err != nil {
		return err
	}
	if !blockExists {
		err := storeBlock(dbTx, block)
		if err != nil {
			return err
		}
	}
	err = dag.index.flushToDB(dbTx)
	if err != nil {
		return err
	}
	err = dbTx.Commit()
	if err != nil {
		return err
	}

	// Make sure that all the block's transactions are finalized
	fastAdd := flags&BFFastAdd == BFFastAdd
	bluestParent := parents.bluest()
	if !fastAdd {
		if err := dag.validateAllTxsFinalized(block, newNode, bluestParent); err != nil {
			return err
		}
	}

	block.SetBlueScore(newNode.blueScore)

	// Connect the passed block to the DAG. This also handles validation of the
	// transaction scripts.
	chainUpdates, err := dag.addBlock(newNode, block, selectedParentAnticone, flags)
	if err != nil {
		return err
	}

	// Notify the caller that the new block was accepted into the block
	// DAG. The caller would typically want to react by relaying the
	// inventory to other peers.
	dag.dagLock.Unlock()
	dag.sendNotification(NTBlockAdded, &BlockAddedNotificationData{
		Block:         block,
		WasUnorphaned: flags&BFWasUnorphaned != 0,
	})
	if len(chainUpdates.addedChainBlockHashes) > 0 {
		dag.sendNotification(NTChainChanged, &ChainChangedNotificationData{
			RemovedChainBlockHashes: chainUpdates.removedChainBlockHashes,
			AddedChainBlockHashes:   chainUpdates.addedChainBlockHashes,
		})
	}
	dag.dagLock.Lock()

	return nil
}

func lookupParentNodes(block *util.Block, blockDAG *BlockDAG) (blockSet, error) {
	header := block.MsgBlock().Header
	parentHashes := header.ParentHashes

	nodes := newBlockSet()
	for _, parentHash := range parentHashes {
		node := blockDAG.index.LookupNode(parentHash)
		if node == nil {
			str := fmt.Sprintf("parent block %s is unknown", parentHash)
			return nil, ruleError(ErrParentBlockUnknown, str)
		} else if blockDAG.index.NodeStatus(node).KnownInvalid() {
			str := fmt.Sprintf("parent block %s is known to be invalid", parentHash)
			return nil, ruleError(ErrInvalidAncestorBlock, str)
		}

		nodes.add(node)
	}

	return nodes, nil
}
