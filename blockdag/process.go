package blockdag

import (
	"fmt"
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/dbaccess"
	"github.com/pkg/errors"
	"time"

	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
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
// When no errors occurred during processing, the first return value indicates
// whether or not the block is an orphan.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) ProcessBlock(block *util.Block, flags BehaviorFlags) (isOrphan bool, isDelayed bool, err error) {
	dag.dagLock.Lock()
	defer dag.dagLock.Unlock()
	return dag.processBlockNoLock(block, flags)
}

func (dag *BlockDAG) processBlockNoLock(block *util.Block, flags BehaviorFlags) (isOrphan bool, isDelayed bool, err error) {
	isAfterDelay := flags&BFAfterDelay == BFAfterDelay
	wasBlockStored := flags&BFWasStored == BFWasStored
	disallowDelay := flags&BFDisallowDelay == BFDisallowDelay
	disallowOrphans := flags&BFDisallowOrphans == BFDisallowOrphans

	blockHash := block.Hash()
	log.Tracef("Processing block %s", blockHash)

	// The block must not already exist in the DAG.
	if dag.IsInDAG(blockHash) && !wasBlockStored {
		str := fmt.Sprintf("already have block %s", blockHash)
		return false, false, ruleError(ErrDuplicateBlock, str)
	}

	// The block must not already exist as an orphan.
	if _, exists := dag.orphans[*blockHash]; exists {
		str := fmt.Sprintf("already have block (orphan) %s", blockHash)
		return false, false, ruleError(ErrDuplicateBlock, str)
	}

	if dag.isKnownDelayedBlock(blockHash) {
		str := fmt.Sprintf("already have block (delayed) %s", blockHash)
		return false, false, ruleError(ErrDuplicateBlock, str)
	}

	if !isAfterDelay {
		// Perform preliminary sanity checks on the block and its transactions.
		delay, err := dag.checkBlockSanity(block, flags)
		if err != nil {
			return false, false, err
		}

		if delay != 0 && disallowDelay {
			str := fmt.Sprintf("Cannot process blocks beyond the allowed time offset while the BFDisallowDelay flag is raised %s", blockHash)
			return false, true, ruleError(ErrDelayedBlockIsNotAllowed, str)
		}

		if delay != 0 {
			err = dag.addDelayedBlock(block, delay)
			if err != nil {
				return false, false, err
			}
			return false, true, nil
		}
	}

	var missingParents []*daghash.Hash
	for _, parentHash := range block.MsgBlock().Header.ParentHashes {
		if !dag.IsInDAG(parentHash) {
			missingParents = append(missingParents, parentHash)
		}
	}
	if len(missingParents) > 0 && disallowOrphans {
		str := fmt.Sprintf("Cannot process orphan blocks while the BFDisallowOrphans flag is raised %s", blockHash)
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
		return false, true, err
	}

	// Handle orphan blocks.
	if len(missingParents) > 0 {
		// Some orphans during netsync are a normal part of the process, since the anticone
		// of the chain-split is never explicitly requested.
		// Therefore, if we are during netsync - don't report orphans to default logs.
		//
		// The number K*2 was chosen since in peace times anticone is limited to K blocks,
		// while some red block can make it a bit bigger, but much more than that indicates
		// there might be some problem with the netsync process.
		if flags&BFIsSync == BFIsSync && dagconfig.KType(len(dag.orphans)) < dag.Params.K*2 {
			log.Debugf("Adding orphan block %s. This is normal part of netsync process", blockHash)
		} else {
			log.Infof("Adding orphan block %s", blockHash)
		}
		dag.addOrphanBlock(block)

		return true, false, nil
	}

	// The block has passed all context independent checks and appears sane
	// enough to potentially accept it into the block DAG.
	err = dag.maybeAcceptBlock(block, flags)
	if err != nil {
		return false, false, err
	}

	// Accept any orphan blocks that depend on this block (they are
	// no longer orphans) and repeat for those accepted blocks until
	// there are no more.
	err = dag.processOrphans(blockHash, flags)
	if err != nil {
		return false, false, err
	}

	if !isAfterDelay {
		err = dag.processDelayedBlocks()
		if err != nil {
			return false, false, err
		}
	}

	dag.addBlockProcessingTimestamp()

	log.Debugf("Accepted block %s", blockHash)

	return false, false, nil
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
	dbTx, err := dag.databaseContext.NewTx()
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
	fastAdd := flags&BFFastAdd == BFFastAdd || dag.index.NodeStatus(newNode).KnownValid()
	bluestParent := parents.bluest()
	if !fastAdd {
		if err := dag.validateAllTxsFinalized(block, newNode, bluestParent); err != nil {
			return err
		}
	}

	// Connect the block to the DAG.
	chainUpdates, err := dag.connectBlock(newNode, block, selectedParentAnticone, fastAdd)
	if err != nil {
		return dag.handleProcessBlockError(err, newNode)
	}
	dag.blockCount++

	dag.notifyBlockAccepted(block, chainUpdates, flags)

	return nil
}

func (dag *BlockDAG) handleProcessBlockError(err error, newNode *blockNode) error {
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
//
// This function assumes that the DAG lock is currently held.
func (dag *BlockDAG) notifyBlockAccepted(block *util.Block, chainUpdates *chainUpdates, flags BehaviorFlags) {
	dag.dagLock.Unlock()
	defer dag.dagLock.Lock()

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

// connectBlock handles connecting the passed node/block to the DAG.
//
// This function MUST be called with the DAG state lock held (for writes).
func (dag *BlockDAG) connectBlock(node *blockNode,
	block *util.Block, selectedParentAnticone []*blockNode, fastAdd bool) (*chainUpdates, error) {

	if err := dag.checkFinalityViolation(node); err != nil {
		return nil, err
	}

	if err := dag.validateGasLimit(block); err != nil {
		return nil, err
	}

	newBlockPastUTXO, txsAcceptanceData, newBlockFeeData, newBlockMultiSet, err :=
		node.verifyAndBuildUTXO(dag, block.Transactions(), fastAdd)
	if err != nil {
		return nil, errors.Wrapf(err, "error verifying UTXO for %s", node)
	}

	err = node.validateCoinbaseTransaction(dag, block, txsAcceptanceData)
	if err != nil {
		return nil, err
	}

	// Apply all changes to the DAG.
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
