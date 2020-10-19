package blockdag

import (
	"fmt"
	"time"

	"github.com/kaspanet/kaspad/domain/utxo"

	"github.com/kaspanet/kaspad/domain/blocknode"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/pkg/errors"
)

// selectedParentChainUpdates represents the updates made to the selected parent chain after
// a block had been added to the DAG.
type selectedParentChainUpdates struct {
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

	isOrphan, isDelayed, err = dag.checkDelayedAndOrphanBlocks(block, flags)
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

	return false, false, nil
}

func (dag *BlockDAG) checkDelayedAndOrphanBlocks(block *util.Block, flags BehaviorFlags) (isOrphan bool, isDelayed bool, err error) {
	if !isBehaviorFlagRaised(flags, BFAfterDelay) {
		isDelayed, err := dag.checkBlockDelay(block, flags)
		if err != nil {
			return false, false, err
		}
		if isDelayed {
			return false, true, nil
		}
	}
	return dag.checkMissingParents(block, flags)
}

func (dag *BlockDAG) checkBlockDelay(block *util.Block, flags BehaviorFlags) (isDelayed bool, err error) {
	delay, isDelayed := dag.shouldBlockBeDelayed(block)
	if isDelayed && isBehaviorFlagRaised(flags, BFDisallowDelay) {
		str := fmt.Sprintf("cannot process blocks beyond the "+
			"allowed time offset while the BFDisallowDelay flag is "+
			"raised %s", block.Hash())
		return false, ruleError(ErrDelayedBlockIsNotAllowed, str)
	}

	if isDelayed {
		err := dag.addDelayedBlock(block, flags, delay)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (dag *BlockDAG) checkMissingParents(block *util.Block, flags BehaviorFlags) (isOrphan bool, isDelayed bool, err error) {
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
		err := dag.addDelayedBlock(block, flags, delay)
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

	log.Debugf("Accepted block %s with status '%s'", newNode.Hash, dag.Index.BlockNodeStatus(newNode))

	return nil
}

// createBlockNodeFromBlock generates a new block node for the given block
// and stores it in the block Index with statusDataStored.
func (dag *BlockDAG) createBlockNodeFromBlock(block *util.Block) (
	newNode *blocknode.Node, selectedParentAnticone []*blocknode.Node, err error) {

	// Create a new block node for the block and add it to the node Index.
	parents, err := lookupParentNodes(block, dag)
	if err != nil {
		return nil, nil, err
	}
	newNode, selectedParentAnticone = dag.newBlockNode(&block.MsgBlock().Header, parents)

	dag.Index.AddNode(newNode)
	dag.Index.SetBlockNodeStatus(newNode, blocknode.StatusDataStored)

	// Insert the block into the database if it's not already there. Even
	// though it is possible the block will ultimately fail to connect, it
	// has already passed all proof-of-work and validity tests which means
	// it would be prohibitively expensive for an attacker to fill up the
	// disk with a bunch of blocks that fail to connect. This is necessary
	// since it allows block download to be decoupled from the much more
	// expensive connection logic. It also has some other nice properties
	// such as making blocks that never become part of the DAG or
	// blocks that fail to connect available for further analysis.
	dbTx, err := dag.DatabaseContext.NewTx()
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
	err = dag.Index.FlushToDB(dbTx)
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
func (dag *BlockDAG) connectBlock(newNode *blocknode.Node,
	block *util.Block, selectedParentAnticone []*blocknode.Node, flags BehaviorFlags) (*selectedParentChainUpdates, error) {

	err := dag.checkDAGRelations(newNode)
	if err != nil {
		return nil, err
	}

	err = dag.checkBlockTransactionsFinalized(block, newNode, flags)
	if err != nil {
		return nil, err
	}

	err = dag.checkBlockHasNoChainedTransactions(block, newNode, flags)
	if err != nil {
		return nil, err
	}

	if err := dag.validateGasLimit(block); err != nil {
		return nil, err
	}

	isNewSelectedTip := dag.isNewSelectedTip(newNode)
	if !isNewSelectedTip {
		dag.Index.SetBlockNodeStatus(newNode, blocknode.StatusUTXOPendingVerification)
	}

	dbTx, err := dag.DatabaseContext.NewTx()
	if err != nil {
		return nil, err
	}
	defer dbTx.RollbackUnlessClosed()

	if isNewSelectedTip {
		err = dag.resolveNodeStatus(newNode, dbTx)
		if err != nil {
			return nil, err
		}
		if dag.Index.BlockNodeStatus(newNode) == blocknode.StatusValid {
			isViolatingFinality, err := dag.isViolatingFinality(newNode)
			if err != nil {
				return nil, err
			}
			if isViolatingFinality {
				dag.Index.SetBlockNodeStatus(newNode, blocknode.StatusUTXOPendingVerification)
				dag.sendNotification(NTFinalityConflict, &FinalityConflictNotificationData{
					ViolatingBlockHash: newNode.Hash,
				})
			}
		}
	}

	chainUpdates, err := dag.applyDAGChanges(newNode, selectedParentAnticone, dbTx)
	if err != nil {
		return nil, err
	}

	err = dag.saveChangesFromBlock(block, dbTx)
	if err != nil {
		return nil, err
	}

	err = dbTx.Commit()
	if err != nil {
		return nil, err
	}

	dag.clearDirtyEntries()

	dag.addBlockProcessingTimestamp()
	dag.blockCount++

	return chainUpdates, nil
}

// isNewSelectedTip determines if a new Node qualifies to be the next selectedTip
func (dag *BlockDAG) isNewSelectedTip(newNode *blocknode.Node) bool {
	return newNode.IsGenesis() || dag.selectedTip().Less(newNode)
}

func (dag *BlockDAG) updateVirtualAndTips(node *blocknode.Node, dbTx *dbaccess.TxContext) (*selectedParentChainUpdates, error) {
	didVirtualParentsChange, chainUpdates, err := dag.addTip(node)
	if err != nil {
		return nil, err
	}

	if didVirtualParentsChange {
		// Build a UTXO set for the new virtual block
		newVirtualUTXO, _, _, err := dag.pastUTXO(dag.virtual.Node)
		if err != nil {
			return nil, errors.Wrap(err, "could not restore past UTXO for virtual")
		}

		// Apply new utxoDiffs to all the tips
		err = updateValidTipsUTXO(dag, newVirtualUTXO)
		if err != nil {
			return nil, errors.Wrap(err, "failed updating the tips' UTXO")
		}

		// It is now safe to meld the UTXO set to base.
		diffSet := newVirtualUTXO.(*utxo.DiffUTXOSet)
		virtualUTXODiff := diffSet.UTXODiff
		err = dag.meldVirtualUTXO(diffSet)
		if err != nil {
			return nil, errors.Wrap(err, "failed melding the virtual UTXO")
		}

		// Update the UTXO set using the diffSet that was melded into the
		// full UTXO set.
		err = utxo.UpdateUTXOSet(dbTx, virtualUTXODiff)
		if err != nil {
			return nil, err
		}
	}
	return chainUpdates, nil
}

func (dag *BlockDAG) validateAndApplyUTXOSet(
	node *blocknode.Node, block *util.Block, dbTx *dbaccess.TxContext) error {

	if !node.IsGenesis() {
		err := dag.resolveNodeStatus(node.SelectedParent, dbTx)
		if err != nil {
			return err
		}

		if dag.Index.BlockNodeStatus(node.SelectedParent) == blocknode.StatusDisqualifiedFromChain {
			return ruleError(ErrSelectedParentDisqualifiedFromChain,
				"Block's selected parent is disqualified from chain")
		}
	}

	utxoVerificationData, err := dag.verifyAndBuildUTXO(node, block.Transactions())
	if err != nil {
		return errors.Wrapf(err, "error verifying UTXO for %s", node)
	}

	err = dag.validateCoinbaseTransaction(node, block, utxoVerificationData.txsAcceptanceData)
	if err != nil {
		return err
	}

	err = dag.applyUTXOSetChanges(node, utxoVerificationData, dbTx)
	if err != nil {
		return err
	}

	return nil
}

func (dag *BlockDAG) applyUTXOSetChanges(
	node *blocknode.Node, utxoVerificationData *utxoVerificationOutput, dbTx *dbaccess.TxContext) error {

	dag.Index.SetBlockNodeStatus(node, blocknode.StatusValid)

	if !dag.hasValidChildren(node) {
		err := dag.addValidTip(node)
		if err != nil {
			return err
		}
	}

	dag.multisetStore.SetMultiset(node, utxoVerificationData.newBlockMultiset)

	err := dag.updateDiffAndDiffChild(node, utxoVerificationData.newBlockPastUTXO)
	if err != nil {
		return err
	}

	if err := dag.updateParentsDiffs(node, utxoVerificationData.newBlockPastUTXO); err != nil {
		return errors.Wrapf(err, "failed updating parents of %s", node)
	}

	if dag.indexManager != nil {
		err := dag.indexManager.ConnectBlock(dbTx, node.Hash, utxoVerificationData.txsAcceptanceData)
		if err != nil {
			return err
		}
	}

	return nil
}

func (dag *BlockDAG) resolveNodeStatus(node *blocknode.Node, dbTx *dbaccess.TxContext) error {
	blockStatus := dag.Index.BlockNodeStatus(node)
	if blockStatus != blocknode.StatusValid && blockStatus != blocknode.StatusDisqualifiedFromChain {
		block, err := dag.fetchBlockByHash(node.Hash)
		if err != nil {
			return err
		}

		err = dag.validateAndApplyUTXOSet(node, block, dbTx)
		if err != nil {
			if !errors.As(err, &(RuleError{})) {
				return err
			}
			dag.Index.SetBlockNodeStatus(node, blocknode.StatusDisqualifiedFromChain)
		}
	}
	return nil
}

func (dag *BlockDAG) resolveNodeStatusInNewTransaction(node *blocknode.Node) error {
	dbTx, err := dag.DatabaseContext.NewTx()
	if err != nil {
		return err
	}
	defer dbTx.RollbackUnlessClosed()
	err = dag.resolveNodeStatus(node, dbTx)
	if err != nil {
		return err
	}
	err = dbTx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (dag *BlockDAG) applyDAGChanges(node *blocknode.Node, selectedParentAnticone []*blocknode.Node, dbTx *dbaccess.TxContext) (
	*selectedParentChainUpdates, error) {

	// Add the block to the reachability tree
	err := dag.reachabilityTree.addBlock(node, selectedParentAnticone)
	if err != nil {
		return nil, errors.Wrap(err, "failed adding block to the reachability tree")
	}

	node.UpdateParentsChildren()

	chainUpdates, err := dag.updateVirtualAndTips(node, dbTx)
	if err != nil {
		return nil, err
	}

	return chainUpdates, nil
}

func (dag *BlockDAG) saveChangesFromBlock(block *util.Block, dbTx *dbaccess.TxContext) error {
	err := dag.Index.FlushToDB(dbTx)
	if err != nil {
		return err
	}

	err = dag.UTXODiffStore.FlushToDB(dbTx)
	if err != nil {
		return err
	}

	err = dag.reachabilityTree.storeState(dbTx)
	if err != nil {
		return err
	}

	err = dag.multisetStore.FlushToDB(dbTx)
	if err != nil {
		return err
	}

	// Update DAG state.
	err = dag.saveState(dbTx)
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

	return nil
}

// boundedMergeBreakingParents returns all parents of given `node` that break the bounded merge depth rule:
// All blocks in node.MergeSet should be in future of node.finalityPoint, with the following exception:
// If there exists a block C violating this, i.e., C is in node's merge set and node.finalityPoint's anticone,
// then there must be a "kosherizing" block D in C's Future such that D is in node.blues
// and node.finalityPoint is in D.SelectedChain
func (dag *BlockDAG) boundedMergeBreakingParents(node *blocknode.Node) (blocknode.Set, error) {
	potentiallyKosherizingBlocks, err := dag.nonBoundedMergeDepthViolatingBlues(node)
	if err != nil {
		return nil, err
	}
	badReds := []*blocknode.Node{}

	finalityPoint := dag.finalityPoint(node)
	for _, redBlock := range node.Reds {
		isFinalityPointInPast, err := dag.isInPast(finalityPoint, redBlock)
		if err != nil {
			return nil, err
		}
		if isFinalityPointInPast {
			continue
		}

		isKosherized := false
		for potentiallyKosherizingBlock := range potentiallyKosherizingBlocks {
			isKosherized, err = dag.isInPast(redBlock, potentiallyKosherizingBlock)
			if err != nil {
				return nil, err
			}
			if isKosherized {
				break
			}
		}

		if !isKosherized {
			badReds = append(badReds, redBlock)
		}
	}

	boundedMergeBreakingParents := blocknode.NewSet()
	for parent := range node.Parents {
		isBadRedInPast := false
		for _, badRedBlock := range badReds {
			isBadRedInPast, err = dag.isInPast(badRedBlock, parent)
			if err != nil {
				return nil, err
			}
			if isBadRedInPast {
				break
			}
		}

		if isBadRedInPast {
			boundedMergeBreakingParents.Add(parent)
		}
	}
	return boundedMergeBreakingParents, nil
}

func (dag *BlockDAG) clearDirtyEntries() {
	dag.Index.ClearDirtyEntries()
	dag.UTXODiffStore.ClearDirtyEntries()
	dag.UTXODiffStore.ClearOldEntries(dag.VirtualBlueScore(), dag.virtual.Parents)
	dag.reachabilityTree.store.clearDirtyEntries()
	dag.multisetStore.ClearNewEntries()
}

func (dag *BlockDAG) handleConnectBlockError(err error, newNode *blocknode.Node) error {
	if errors.As(err, &RuleError{}) {
		dag.Index.SetBlockNodeStatus(newNode, blocknode.StatusValidateFailed)

		dbTx, err := dag.DatabaseContext.NewTx()
		if err != nil {
			return err
		}
		defer dbTx.RollbackUnlessClosed()

		err = dag.Index.FlushToDB(dbTx)
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
func (dag *BlockDAG) notifyBlockAccepted(block *util.Block, chainUpdates *selectedParentChainUpdates, flags BehaviorFlags) {
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
