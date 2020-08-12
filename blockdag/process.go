package blockdag

import (
	"fmt"
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/dbaccess"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/pkg/errors"
	"time"
)

// chainUpdates represents the updates made to the selected parent chain after
// a block had been added to the DAG.
type chainUpdates struct {
	removedChainBlockHashes []*daghash.Hash
	addedChainBlockHashes   []*daghash.Hash
}

// ProcessBlock is the main workhorse for handling insertion of new blocks into
// the block DAG. It includes functionality such as rejecting duplicate
// blocks, ensuring blocks follow all rules, orphan handling, and insertion into
// the block DAG.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) ProcessBlock(block *util.Block, flags BehaviorFlags) (isOrphan bool, isDelayed bool, err error) {
	dag.dagLock.Lock()
	defer dag.dagLock.Unlock()
	return dag.processBlockNoLock(block, flags)
}

func (dag *BlockDAG) processBlockNoLock(block *util.Block, flags BehaviorFlags) (isOrphan bool, isDelayed bool, err error) {
	blockHash := block.Hash()
	log.Tracef("Processing block %s", blockHash)

	err = dag.checkDuplicateBlock(blockHash, flags)
	if err != nil {
		return false, false, err
	}

	err = dag.checkBlockSanity(block, flags)
	if err != nil {
		return false, false, err
	}

	isOrphan, isDelayed, err = dag.processBlockCheckDelayedAndOrphanBlocks(block, flags)
	if isOrphan || isDelayed || err != nil {
		return isOrphan, isDelayed, err
	}

	err = dag.maybeAcceptBlock(block, flags)
	if err != nil {
		return false, false, err
	}

	err = dag.processOrphansAndDelayedBlocks(blockHash, flags)
	if err != nil {
		return false, false, err
	}

	log.Debugf("Accepted block %s", blockHash)

	return false, false, nil
}

func (dag *BlockDAG) processBlockCheckDelayedAndOrphanBlocks(block *util.Block, flags BehaviorFlags) (isOrphan bool, isDelayed bool, err error) {
	if !isBehaviorFlagRaised(flags, BFAfterDelay) {
		isDelayed, err := dag.processBlockCheckBlockDelay(block, flags)
		if err != nil {
			return false, false, err
		}
		if isDelayed {
			return false, true, nil
		}
	}
	return dag.processBlockCheckMissingParents(block, flags)
}

func (dag *BlockDAG) processBlockCheckBlockDelay(block *util.Block, flags BehaviorFlags) (isDelayed bool, err error) {
	delay, isDelayed := dag.checkBlockDelayed(block)
	if isDelayed && isBehaviorFlagRaised(flags, BFDisallowDelay) {
		str := fmt.Sprintf("cannot process blocks beyond the "+
			"allowed time offset while the BFDisallowDelay flag is "+
			"raised %s", block.Hash())
		return false, ruleError(ErrDelayedBlockIsNotAllowed, str)
	}

	if isDelayed {
		err := dag.addDelayedBlock(block, delay)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (dag *BlockDAG) processBlockCheckMissingParents(block *util.Block, flags BehaviorFlags) (isOrphan bool, isDelayed bool, err error) {
	var missingParents []*daghash.Hash
	for _, parentHash := range block.MsgBlock().Header.ParentHashes {
		if !dag.IsInDAG(parentHash) {
			missingParents = append(missingParents, parentHash)
		}
	}

	if len(missingParents) > 0 && isBehaviorFlagRaised(flags, BFDisallowOrphans) {
		str := fmt.Sprintf("cannot process orphan blocks while the "+
			"BFDisallowOrphans flag is raised %s", block.Hash())
		return false, false, ruleError(ErrOrphanBlockIsNotAllowed, str)
	}

	// Handle the case of a block with a valid timestamp(non-delayed) which points to a delayed block.
	delay, isParentDelayed := dag.maxDelayOfParents(missingParents)
	if isParentDelayed {
		// Add Millisecond to ensure that parent process time will be after its child.
		delay += time.Millisecond
		err := dag.addDelayedBlock(block, delay)
		if err != nil {
			return false, false, err
		}
		return false, true, nil
	}

	// Handle orphan blocks.
	if len(missingParents) > 0 {
		dag.addOrphanBlock(block)
		return true, false, nil
	}
	return false, false, nil
}

func (dag *BlockDAG) processOrphansAndDelayedBlocks(blockHash *daghash.Hash, flags BehaviorFlags) error {
	err := dag.processOrphans(blockHash, flags)
	if err != nil {
		return err
	}

	if !isBehaviorFlagRaised(flags, BFAfterDelay) {
		err = dag.processDelayedBlocks()
		if err != nil {
			return err
		}
	}
	return nil
}

// maybeAcceptBlock potentially accepts a block into the block DAG. It
// performs several validation checks which depend on its position within
// the block DAG before adding it. The block is expected to have already
// gone through ProcessBlock before calling this function with it.
//
// The flags are also passed to checkBlockContext and connectBlock. See
// their documentation for how the flags modify their behavior.
//
// This function MUST be called with the dagLock held (for writes).
func (dag *BlockDAG) maybeAcceptBlock(block *util.Block, flags BehaviorFlags) error {
	err := dag.checkBlockContext(block, flags)
	if err != nil {
		return err
	}

	newNode, selectedParentAnticone, err := dag.createBlockNodeFromBlock(block)
	if err != nil {
		return err
	}

	chainUpdates, err := dag.connectBlock(newNode, block, selectedParentAnticone, flags)
	if err != nil {
		return dag.handleConnectBlockError(err, newNode)
	}

	dag.notifyBlockAccepted(block, chainUpdates, flags)

	return nil
}

func (dag *BlockDAG) createBlockNodeFromBlock(block *util.Block) (
	newNode *blockNode, selectedParentAnticone []*blockNode, err error) {

	// Create a new block node for the block and add it to the node index.
	parents, err := lookupParentNodes(block, dag)
	if err != nil {
		return nil, nil, err
	}
	newNode, selectedParentAnticone = dag.newBlockNode(&block.MsgBlock().Header, parents)
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
	dbTx, err := dag.databaseContext.NewTx()
	if err != nil {
		return nil, nil, err
	}
	defer dbTx.RollbackUnlessClosed()
	blockExists, err := dbaccess.HasBlock(dbTx, block.Hash())
	if err != nil {
		return nil, nil, err
	}
	if !blockExists {
		err := storeBlock(dbTx, block)
		if err != nil {
			return nil, nil, err
		}
	}
	err = dag.index.flushToDB(dbTx)
	if err != nil {
		return nil, nil, err
	}
	err = dbTx.Commit()
	if err != nil {
		return nil, nil, err
	}
	return newNode, selectedParentAnticone, nil
}

// connectBlock handles connecting the passed node/block to the DAG.
//
// This function MUST be called with the DAG state lock held (for writes).
func (dag *BlockDAG) connectBlock(node *blockNode,
	block *util.Block, selectedParentAnticone []*blockNode, flags BehaviorFlags) (*chainUpdates, error) {

	err := dag.checkBlockTransactionsFinalized(block, node, flags)
	if err != nil {
		return nil, err
	}

	if err = dag.checkFinalityViolation(node); err != nil {
		return nil, err
	}

	if err := dag.validateGasLimit(block); err != nil {
		return nil, err
	}

	newBlockPastUTXO, txsAcceptanceData, newBlockFeeData, newBlockMultiSet, err :=
		node.verifyAndBuildUTXO(dag, block.Transactions(), isBehaviorFlagRaised(flags, BFFastAdd))
	if err != nil {
		return nil, errors.Wrapf(err, "error verifying UTXO for %s", node)
	}

	err = node.validateCoinbaseTransaction(dag, block, txsAcceptanceData)
	if err != nil {
		return nil, err
	}

	virtualUTXODiff, chainUpdates, err :=
		dag.applyDAGChanges(node, newBlockPastUTXO, newBlockMultiSet, selectedParentAnticone)
	if err != nil {
		// Since all validation logic has already ran, if applyDAGChanges errors out,
		// this means we have a problem in the internal structure of the DAG - a problem which is
		// irrecoverable, and it would be a bad idea to attempt adding any more blocks to the DAG.
		// Therefore - in such cases we panic.
		panic(err)
	}

	err = dag.saveChangesFromBlock(block, virtualUTXODiff, txsAcceptanceData, newBlockFeeData)
	if err != nil {
		return nil, err
	}

	dag.addBlockProcessingTimestamp()
	dag.blockCount++

	return chainUpdates, nil
}

// applyDAGChanges does the following:
// 1. Connects each of the new block's parents to the block.
// 2. Adds the new block to the DAG's tips.
// 3. Updates the DAG's full UTXO set.
// 4. Updates each of the tips' utxoDiff.
// 5. Applies the new virtual's blue score to all the unaccepted UTXOs
// 6. Adds the block to the reachability structures
// 7. Adds the multiset of the block to the multiset store.
// 8. Updates the finality point of the DAG (if required).
//
// It returns the diff in the virtual block's UTXO set.
//
// This function MUST be called with the DAG state lock held (for writes).
func (dag *BlockDAG) applyDAGChanges(node *blockNode, newBlockPastUTXO UTXOSet,
	newBlockMultiset *secp256k1.MultiSet, selectedParentAnticone []*blockNode) (
	virtualUTXODiff *UTXODiff, chainUpdates *chainUpdates, err error) {

	// Add the block to the reachability tree
	err = dag.reachabilityTree.addBlock(node, selectedParentAnticone)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed adding block to the reachability tree")
	}

	dag.multisetStore.setMultiset(node, newBlockMultiset)

	if err = node.updateParents(dag, newBlockPastUTXO); err != nil {
		return nil, nil, errors.Wrapf(err, "failed updating parents of %s", node)
	}

	// Update the virtual block's parents (the DAG tips) to include the new block.
	chainUpdates = dag.virtual.AddTip(node)

	// Build a UTXO set for the new virtual block
	newVirtualUTXO, _, _, err := dag.pastUTXO(&dag.virtual.blockNode)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not restore past UTXO for virtual")
	}

	// Apply new utxoDiffs to all the tips
	err = updateTipsUTXO(dag, newVirtualUTXO)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed updating the tips' UTXO")
	}

	// It is now safe to meld the UTXO set to base.
	diffSet := newVirtualUTXO.(*DiffUTXOSet)
	virtualUTXODiff = diffSet.UTXODiff
	err = dag.meldVirtualUTXO(diffSet)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed melding the virtual UTXO")
	}

	dag.index.SetStatusFlags(node, statusValid)

	// And now we can update the finality point of the DAG (if required)
	dag.updateFinalityPoint()

	return virtualUTXODiff, chainUpdates, nil
}

func (dag *BlockDAG) saveChangesFromBlock(block *util.Block, virtualUTXODiff *UTXODiff,
	txsAcceptanceData MultiBlockTxsAcceptanceData, feeData compactFeeData) error {

	dbTx, err := dag.databaseContext.NewTx()
	if err != nil {
		return err
	}
	defer dbTx.RollbackUnlessClosed()

	err = dag.index.flushToDB(dbTx)
	if err != nil {
		return err
	}

	err = dag.utxoDiffStore.flushToDB(dbTx)
	if err != nil {
		return err
	}

	err = dag.reachabilityTree.storeState(dbTx)
	if err != nil {
		return err
	}

	err = dag.multisetStore.flushToDB(dbTx)
	if err != nil {
		return err
	}

	// Update DAG state.
	state := &dagState{
		TipHashes:         dag.TipHashes(),
		LastFinalityPoint: dag.lastFinalityPoint.hash,
		LocalSubnetworkID: dag.subnetworkID,
	}
	err = saveDAGState(dbTx, state)
	if err != nil {
		return err
	}

	// Update the UTXO set using the diffSet that was melded into the
	// full UTXO set.
	err = updateUTXOSet(dbTx, virtualUTXODiff)
	if err != nil {
		return err
	}

	// Scan all accepted transactions and register any subnetwork registry
	// transaction. If any subnetwork registry transaction is not well-formed,
	// fail the entire block.
	err = registerSubnetworks(dbTx, block.Transactions())
	if err != nil {
		return err
	}

	// Allow the index manager to call each of the currently active
	// optional indexes with the block being connected so they can
	// update themselves accordingly.
	if dag.indexManager != nil {
		err := dag.indexManager.ConnectBlock(dbTx, block.Hash(), txsAcceptanceData)
		if err != nil {
			return err
		}
	}

	// Apply the fee data into the database
	err = dbaccess.StoreFeeData(dbTx, block.Hash(), feeData)
	if err != nil {
		return err
	}

	err = dbTx.Commit()
	if err != nil {
		return err
	}

	dag.index.clearDirtyEntries()
	dag.utxoDiffStore.clearDirtyEntries()
	dag.utxoDiffStore.clearOldEntries()
	dag.reachabilityTree.store.clearDirtyEntries()
	dag.multisetStore.clearNewEntries()

	return nil
}

func (dag *BlockDAG) handleConnectBlockError(err error, newNode *blockNode) error {
	if errors.As(err, &RuleError{}) {
		dag.index.SetStatusFlags(newNode, statusValidateFailed)

		dbTx, err := dag.databaseContext.NewTx()
		if err != nil {
			return err
		}
		defer dbTx.RollbackUnlessClosed()

		err = dag.index.flushToDB(dbTx)
		if err != nil {
			return err
		}
		err = dbTx.Commit()
		if err != nil {
			return err
		}
	}
	return err
}

// notifyBlockAccepted notifies the caller that the new block was
// accepted into the block DAG. The caller would typically want to
// react by relaying the inventory to other peers.
func (dag *BlockDAG) notifyBlockAccepted(block *util.Block, chainUpdates *chainUpdates, flags BehaviorFlags) {
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
}
