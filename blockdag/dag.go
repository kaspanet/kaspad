// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/daglabs/btcd/util/subnetworkid"

	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/wire"
)

const (
	// maxOrphanBlocks is the maximum number of orphan blocks that can be
	// queued.
	maxOrphanBlocks = 100

	// FinalityInterval is the interval that determines the finality window of the DAG.
	FinalityInterval = 100
)

// BlockLocator is used to help locate a specific block.  The algorithm for
// building the block locator is to add the hashes in reverse order until
// the genesis block is reached.  In order to keep the list of locator hashes
// to a reasonable number of entries, first the most recent previous 12 block
// hashes are added, then the step is doubled each loop iteration to
// exponentially decrease the number of hashes as a function of the distance
// from the block being located.
//
// For example, assume a block chain with a side chain as depicted below:
// 	genesis -> 1 -> 2 -> ... -> 15 -> 16  -> 17  -> 18
// 	                              \-> 16a -> 17a
//
// The block locator for block 17a would be the hashes of blocks:
// [17a 16a 15 14 13 12 11 10 9 8 7 6 4 genesis]
type BlockLocator []*daghash.Hash

// orphanBlock represents a block that we don't yet have the parent for.  It
// is a normal block plus an expiration time to prevent caching the orphan
// forever.
type orphanBlock struct {
	block      *util.Block
	expiration time.Time
}

// BlockDAG provides functions for working with the bitcoin block chain.
// It includes functionality such as rejecting duplicate blocks, ensuring blocks
// follow all rules, orphan handling, checkpoint handling, and best chain
// selection with reorganization.
type BlockDAG struct {
	// The following fields are set when the instance is created and can't
	// be changed afterwards, so there is no need to protect them with a
	// separate mutex.
	checkpoints         []dagconfig.Checkpoint
	checkpointsByHeight map[int32]*dagconfig.Checkpoint
	db                  database.DB
	dagParams           *dagconfig.Params
	timeSource          MedianTimeSource
	sigCache            *txscript.SigCache
	indexManager        IndexManager
	genesis             *blockNode

	// The following fields are calculated based upon the provided chain
	// parameters.  They are also set when the instance is created and
	// can't be changed afterwards, so there is no need to protect them with
	// a separate mutex.
	minRetargetTimespan int64 // target timespan / adjustment factor
	maxRetargetTimespan int64 // target timespan * adjustment factor
	blocksPerRetarget   int32 // target timespan / target time per block

	// dagLock protects concurrent access to the vast majority of the
	// fields in this struct below this point.
	dagLock sync.RWMutex

	utxoLock sync.RWMutex

	// index and virtual are related to the memory block index.  They both
	// have their own locks, however they are often also protected by the
	// DAG lock to help prevent logic races when blocks are being processed.

	// index houses the entire block index in memory.  The block index is
	// a tree-shaped structure.
	index *blockIndex

	// blockCount holds the number of blocks in the DAG
	blockCount uint64

	// virtual tracks the current tips.
	virtual *virtualBlock

	// subnetworkID holds the subnetwork ID of the DAG
	subnetworkID *subnetworkid.SubnetworkID

	// These fields are related to handling of orphan blocks.  They are
	// protected by a combination of the chain lock and the orphan lock.
	orphanLock   sync.RWMutex
	orphans      map[daghash.Hash]*orphanBlock
	prevOrphans  map[daghash.Hash][]*orphanBlock
	oldestOrphan *orphanBlock

	// These fields are related to checkpoint handling.  They are protected
	// by the chain lock.
	nextCheckpoint *dagconfig.Checkpoint
	checkpointNode *blockNode

	// The following caches are used to efficiently keep track of the
	// current deployment threshold state of each rule change deployment.
	//
	// This information is stored in the database so it can be quickly
	// reconstructed on load.
	//
	// warningCaches caches the current deployment threshold state for blocks
	// in each of the **possible** deployments.  This is used in order to
	// detect when new unrecognized rule changes are being voted on and/or
	// have been activated such as will be the case when older versions of
	// the software are being used
	//
	// deploymentCaches caches the current deployment threshold state for
	// blocks in each of the actively defined deployments.
	warningCaches    []thresholdStateCache
	deploymentCaches []thresholdStateCache

	// The following fields are used to determine if certain warnings have
	// already been shown.
	//
	// unknownRulesWarned refers to warnings due to unknown rules being
	// activated.
	//
	// unknownVersionsWarned refers to warnings due to unknown versions
	// being mined.
	unknownRulesWarned    bool
	unknownVersionsWarned bool

	// The notifications field stores a slice of callbacks to be executed on
	// certain blockchain events.
	notificationsLock sync.RWMutex
	notifications     []NotificationCallback

	lastFinalityPoint *blockNode

	SubnetworkStore *SubnetworkStore
}

// HaveBlock returns whether or not the DAG instance has the block represented
// by the passed hash.  This includes checking the various places a block can
// be in, like part of the DAG or the orphan pool.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) HaveBlock(hash *daghash.Hash) (bool, error) {
	exists, err := dag.blockExists(hash)
	if err != nil {
		return false, err
	}
	return exists || dag.IsKnownOrphan(hash), nil
}

// HaveBlocks returns whether or not the DAG instances has all blocks represented
// by the passed hashes. This includes checking the various places a block can
// be in, like part of the DAG or the orphan pool.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) HaveBlocks(hashes []daghash.Hash) (bool, error) {
	for _, hash := range hashes {
		haveBlock, err := dag.HaveBlock(&hash)
		if err != nil {
			return false, err
		}
		if !haveBlock {
			return false, nil
		}
	}

	return true, nil
}

// IsKnownOrphan returns whether the passed hash is currently a known orphan.
// Keep in mind that only a limited number of orphans are held onto for a
// limited amount of time, so this function must not be used as an absolute
// way to test if a block is an orphan block.  A full block (as opposed to just
// its hash) must be passed to ProcessBlock for that purpose.  However, calling
// ProcessBlock with an orphan that already exists results in an error, so this
// function provides a mechanism for a caller to intelligently detect *recent*
// duplicate orphans and react accordingly.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) IsKnownOrphan(hash *daghash.Hash) bool {
	// Protect concurrent access.  Using a read lock only so multiple
	// readers can query without blocking each other.
	dag.orphanLock.RLock()
	_, exists := dag.orphans[*hash]
	dag.orphanLock.RUnlock()

	return exists
}

// GetOrphanRoot returns the head of the chain for the provided hash from the
// map of orphan blocks.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) GetOrphanRoot(hash *daghash.Hash) *daghash.Hash {
	// Protect concurrent access.  Using a read lock only so multiple
	// readers can query without blocking each other.
	dag.orphanLock.RLock()
	defer dag.orphanLock.RUnlock()

	// Keep looping while the parent of each orphaned block is
	// known and is an orphan itself.
	orphanRoot := hash
	parentHash := hash
	for {
		orphan, exists := dag.orphans[*parentHash]
		if !exists {
			break
		}
		orphanRoot = parentHash
		parentHash = orphan.block.MsgBlock().Header.SelectedParentHash()
	}

	return orphanRoot
}

// removeOrphanBlock removes the passed orphan block from the orphan pool and
// previous orphan index.
func (dag *BlockDAG) removeOrphanBlock(orphan *orphanBlock) {
	// Protect concurrent access.
	dag.orphanLock.Lock()
	defer dag.orphanLock.Unlock()

	// Remove the orphan block from the orphan pool.
	orphanHash := orphan.block.Hash()
	delete(dag.orphans, *orphanHash)

	// Remove the reference from the previous orphan index too.  An indexing
	// for loop is intentionally used over a range here as range does not
	// reevaluate the slice on each iteration nor does it adjust the index
	// for the modified slice.
	parentHash := orphan.block.MsgBlock().Header.SelectedParentHash()
	orphans := dag.prevOrphans[*parentHash]
	for i := 0; i < len(orphans); i++ {
		hash := orphans[i].block.Hash()
		if hash.IsEqual(orphanHash) {
			copy(orphans[i:], orphans[i+1:])
			orphans[len(orphans)-1] = nil
			orphans = orphans[:len(orphans)-1]
			i--
		}
	}
	dag.prevOrphans[*parentHash] = orphans

	// Remove the map entry altogether if there are no longer any orphans
	// which depend on the parent hash.
	if len(dag.prevOrphans[*parentHash]) == 0 {
		delete(dag.prevOrphans, *parentHash)
	}
}

// addOrphanBlock adds the passed block (which is already determined to be
// an orphan prior calling this function) to the orphan pool.  It lazily cleans
// up any expired blocks so a separate cleanup poller doesn't need to be run.
// It also imposes a maximum limit on the number of outstanding orphan
// blocks and will remove the oldest received orphan block if the limit is
// exceeded.
func (dag *BlockDAG) addOrphanBlock(block *util.Block) {
	// Remove expired orphan blocks.
	for _, oBlock := range dag.orphans {
		if time.Now().After(oBlock.expiration) {
			dag.removeOrphanBlock(oBlock)
			continue
		}

		// Update the oldest orphan block pointer so it can be discarded
		// in case the orphan pool fills up.
		if dag.oldestOrphan == nil || oBlock.expiration.Before(dag.oldestOrphan.expiration) {
			dag.oldestOrphan = oBlock
		}
	}

	// Limit orphan blocks to prevent memory exhaustion.
	if len(dag.orphans)+1 > maxOrphanBlocks {
		// Remove the oldest orphan to make room for the new one.
		dag.removeOrphanBlock(dag.oldestOrphan)
		dag.oldestOrphan = nil
	}

	// Protect concurrent access.  This is intentionally done here instead
	// of near the top since removeOrphanBlock does its own locking and
	// the range iterator is not invalidated by removing map entries.
	dag.orphanLock.Lock()
	defer dag.orphanLock.Unlock()

	// Insert the block into the orphan map with an expiration time
	// 1 hour from now.
	expiration := time.Now().Add(time.Hour)
	oBlock := &orphanBlock{
		block:      block,
		expiration: expiration,
	}
	dag.orphans[*block.Hash()] = oBlock

	// Add to parent hash lookup index for faster dependency lookups.
	parentHash := block.MsgBlock().Header.SelectedParentHash()
	dag.prevOrphans[*parentHash] = append(dag.prevOrphans[*parentHash], oBlock)
}

// SequenceLock represents the converted relative lock-time in seconds, and
// absolute block-height for a transaction input's relative lock-times.
// According to SequenceLock, after the referenced input has been confirmed
// within a block, a transaction spending that input can be included into a
// block either after 'seconds' (according to past median time), or once the
// 'BlockHeight' has been reached.
type SequenceLock struct {
	Seconds     int64
	BlockHeight int32
}

// CalcSequenceLock computes a relative lock-time SequenceLock for the passed
// transaction using the passed UTXOSet to obtain the past median time
// for blocks in which the referenced inputs of the transactions were included
// within. The generated SequenceLock lock can be used in conjunction with a
// block height, and adjusted median block time to determine if all the inputs
// referenced within a transaction have reached sufficient maturity allowing
// the candidate transaction to be included in a block.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) CalcSequenceLock(tx *util.Tx, utxoSet UTXOSet, mempool bool) (*SequenceLock, error) {
	dag.dagLock.Lock()
	defer dag.dagLock.Unlock()

	return dag.calcSequenceLock(dag.selectedTip(), utxoSet, tx, mempool)
}

// calcSequenceLock computes the relative lock-times for the passed
// transaction. See the exported version, CalcSequenceLock for further details.
//
// This function MUST be called with the DAG state lock held (for writes).
func (dag *BlockDAG) calcSequenceLock(node *blockNode, utxoSet UTXOSet, tx *util.Tx, mempool bool) (*SequenceLock, error) {
	// A value of -1 for each relative lock type represents a relative time
	// lock value that will allow a transaction to be included in a block
	// at any given height or time.
	sequenceLock := &SequenceLock{Seconds: -1, BlockHeight: -1}

	// Sequence locks don't apply to block reward transactions Therefore, we
	// return sequence lock values of -1 indicating that this transaction
	// can be included within a block at any given height or time.
	if IsBlockReward(tx) {
		return sequenceLock, nil
	}

	// Grab the next height from the PoV of the passed blockNode to use for
	// inputs present in the mempool.
	nextHeight := node.height + 1

	mTx := tx.MsgTx()
	for txInIndex, txIn := range mTx.TxIn {
		entry, ok := utxoSet.Get(txIn.PreviousOutPoint)
		if !ok {
			str := fmt.Sprintf("output %s referenced from "+
				"transaction %s:%d either does not exist or "+
				"has already been spent", txIn.PreviousOutPoint,
				tx.ID(), txInIndex)
			return sequenceLock, ruleError(ErrMissingTxOut, str)
		}

		// If the input height is set to the mempool height, then we
		// assume the transaction makes it into the next block when
		// evaluating its sequence blocks.
		inputHeight := entry.BlockHeight()
		if inputHeight == 0x7fffffff {
			inputHeight = nextHeight
		}

		// Given a sequence number, we apply the relative time lock
		// mask in order to obtain the time lock delta required before
		// this input can be spent.
		sequenceNum := txIn.Sequence
		relativeLock := int64(sequenceNum & wire.SequenceLockTimeMask)

		switch {
		// Relative time locks are disabled for this input, so we can
		// skip any further calculation.
		case sequenceNum&wire.SequenceLockTimeDisabled == wire.SequenceLockTimeDisabled:
			continue
		case sequenceNum&wire.SequenceLockTimeIsSeconds == wire.SequenceLockTimeIsSeconds:
			// This input requires a relative time lock expressed
			// in seconds before it can be spent.  Therefore, we
			// need to query for the block prior to the one in
			// which this input was included within so we can
			// compute the past median time for the block prior to
			// the one which included this referenced output.
			prevInputHeight := inputHeight - 1
			if prevInputHeight < 0 {
				prevInputHeight = 0
			}
			blockNode := node.SelectedAncestor(prevInputHeight)
			medianTime := blockNode.PastMedianTime()

			// Time based relative time-locks as defined by BIP 68
			// have a time granularity of RelativeLockSeconds, so
			// we shift left by this amount to convert to the
			// proper relative time-lock. We also subtract one from
			// the relative lock to maintain the original lockTime
			// semantics.
			timeLockSeconds := (relativeLock << wire.SequenceLockTimeGranularity) - 1
			timeLock := medianTime.Unix() + timeLockSeconds
			if timeLock > sequenceLock.Seconds {
				sequenceLock.Seconds = timeLock
			}
		default:
			// The relative lock-time for this input is expressed
			// in blocks so we calculate the relative offset from
			// the input's height as its converted absolute
			// lock-time. We subtract one from the relative lock in
			// order to maintain the original lockTime semantics.
			blockHeight := inputHeight + int32(relativeLock-1)
			if blockHeight > sequenceLock.BlockHeight {
				sequenceLock.BlockHeight = blockHeight
			}
		}
	}

	return sequenceLock, nil
}

// LockTimeToSequence converts the passed relative locktime to a sequence
// number in accordance to BIP-68.
// See: https://github.com/bitcoin/bips/blob/master/bip-0068.mediawiki
//  * (Compatibility)
func LockTimeToSequence(isSeconds bool, locktime uint64) uint64 {
	// If we're expressing the relative lock time in blocks, then the
	// corresponding sequence number is simply the desired input age.
	if !isSeconds {
		return locktime
	}

	// Set the 22nd bit which indicates the lock time is in seconds, then
	// shift the locktime over by 9 since the time granularity is in
	// 512-second intervals (2^9). This results in a max lock-time of
	// 33,553,920 seconds, or 1.1 years.
	return wire.SequenceLockTimeIsSeconds |
		locktime>>wire.SequenceLockTimeGranularity
}

// addBlock handles adding the passed block to the DAG.
//
// The flags modify the behavior of this function as follows:
//  - BFFastAdd: Avoids several expensive transaction validation operations.
//    This is useful when using checkpoints.
//
// This function MUST be called with the DAG state lock held (for writes).
func (dag *BlockDAG) addBlock(node *blockNode, parentNodes blockSet, block *util.Block, flags BehaviorFlags) error {
	// Skip checks if node has already been fully validated.
	fastAdd := flags&BFFastAdd == BFFastAdd || dag.index.NodeStatus(node).KnownValid()

	// Connect the block to the DAG.
	err := dag.connectBlock(node, block, fastAdd)
	if err != nil {
		if _, ok := err.(RuleError); ok {
			dag.index.SetStatusFlags(node, statusValidateFailed)
		} else {
			return err
		}
	} else {
		dag.blockCount++
	}

	if err != nil {
		// Notify the caller that the block was connected to the DAG.
		// The caller would typically want to react with actions such as
		// updating wallets.
		dag.dagLock.Unlock()
		dag.sendNotification(NTBlockConnected, block)
		dag.dagLock.Lock()
	}

	// Intentionally ignore errors writing updated node status to DB. If
	// it fails to write, it's not the end of the world. If the block is
	// invalid, the worst that can happen is we revalidate the block
	// after a restart.
	if writeErr := dag.index.flushToDB(); writeErr != nil {
		log.Warnf("Error flushing block index changes to disk: %s",
			writeErr)
	}

	// If dag.connectBlock returned a rule error, return it here after updating DB
	return err
}

// connectBlock handles connecting the passed node/block to the DAG.
//
// This function MUST be called with the DAG state lock held (for writes).
func (dag *BlockDAG) connectBlock(node *blockNode, block *util.Block, fastAdd bool) error {
	// No warnings about unknown rules or versions until the DAG is
	// current.
	if dag.isCurrent() {
		// Warn if any unknown new rules are either about to activate or
		// have already been activated.
		if err := dag.warnUnknownRuleActivations(node); err != nil {
			return err
		}

		// Warn if a high enough percentage of the last blocks have
		// unexpected versions.
		if err := dag.warnUnknownVersions(node); err != nil {
			return err
		}
	}

	if err := dag.checkFinalityRules(node); err != nil {
		return err
	}

	if err := dag.validateGasLimit(block); err != nil {
		return err
	}

	newBlockUTXO, txsAcceptanceData, newBlockFeeData, err := node.verifyAndBuildUTXO(dag, block.Transactions(), fastAdd)
	if err != nil {
		newErrString := fmt.Sprintf("error verifying UTXO for %s: %s", node, err)
		if err, ok := err.(RuleError); ok {
			return ruleError(err.ErrorCode, newErrString)
		}
		return errors.New(newErrString)
	}

	err = node.validateFeeTransaction(dag, block, txsAcceptanceData)
	if err != nil {
		return err
	}

	// Apply all changes to the DAG.
	virtualUTXODiff, err := dag.applyDAGChanges(node, block, newBlockUTXO, fastAdd)
	if err != nil {
		// Since all validation logic has already ran, if applyDAGChanges errors out,
		// this means we have a problem in the internal structure of the DAG - a problem which is
		// irrecoverable, and it would be a bad idea to attempt adding any more blocks to the DAG.
		// Therefore - in such cases we panic.
		panic(err)
	}

	// Write any block status changes to DB before updating the DAG state.
	err = dag.index.flushToDB()
	if err != nil {
		return err
	}

	err = dag.saveChangesFromBlock(node, block, virtualUTXODiff, txsAcceptanceData, newBlockFeeData)
	if err != nil {
		return err
	}

	return nil
}

func (dag *BlockDAG) saveChangesFromBlock(node *blockNode, block *util.Block, virtualUTXODiff *UTXODiff,
	txsAcceptanceData MultiblockTxsAcceptanceData, feeData compactFeeData) error {

	// Atomically insert info into the database.
	return dag.db.Update(func(dbTx database.Tx) error {
		// Update best block state.
		state := &dagState{
			TipHashes:         dag.TipHashes(),
			LastFinalityPoint: dag.lastFinalityPoint.hash,
		}
		err := dbPutDAGState(dbTx, state)
		if err != nil {
			return err
		}

		// Add the block hash and height to the block index which tracks
		// the main chain.
		err = dbPutBlockIndex(dbTx, block.Hash(), node.height)
		if err != nil {
			return err
		}

		// Update the UTXO set using the diffSet that was melded into the
		// full UTXO set.
		err = dbPutUTXODiff(dbTx, virtualUTXODiff)
		if err != nil {
			return err
		}

		// Scan all accepted transactions and register any subnetwork registry
		// transaction. If any subnetwork registry transaction is not well-formed,
		// fail the entire block.
		err = registerSubnetworks(dbTx, txsAcceptanceData)
		if err != nil {
			return err
		}

		// Allow the index manager to call each of the currently active
		// optional indexes with the block being connected so they can
		// update themselves accordingly.
		if dag.indexManager != nil {
			err := dag.indexManager.ConnectBlock(dbTx, block, dag, txsAcceptanceData)
			if err != nil {
				return err
			}
		}

		// Apply the fee data into the database
		return dbStoreFeeData(dbTx, block.Hash(), feeData)
	})
}

func (dag *BlockDAG) validateGasLimit(block *util.Block) error {
	transactions := block.Transactions()
	// Amount of gas consumed per sub-network shouldn't be more than the subnetwork's limit
	gasUsageInAllSubnetworks := map[subnetworkid.SubnetworkID]uint64{}
	for _, tx := range transactions {
		msgTx := tx.MsgTx()
		// In DAGCoin and Registry sub-networks all txs must have Gas = 0, and that is validated in checkTransactionSanity
		// Therefore - no need to check them here.
		if !msgTx.SubnetworkID.IsEqual(subnetworkid.SubnetworkIDNative) && !msgTx.SubnetworkID.IsEqual(subnetworkid.SubnetworkIDRegistry) {
			gasUsageInSubnetwork := gasUsageInAllSubnetworks[msgTx.SubnetworkID]
			gasUsageInSubnetwork += msgTx.Gas
			if gasUsageInSubnetwork < gasUsageInAllSubnetworks[msgTx.SubnetworkID] { // protect from overflows
				str := fmt.Sprintf("Block gas usage in subnetwork with ID %s has overflown", msgTx.SubnetworkID)
				return ruleError(ErrInvalidGas, str)
			}
			gasUsageInAllSubnetworks[msgTx.SubnetworkID] = gasUsageInSubnetwork

			gasLimit, err := dag.SubnetworkStore.GasLimit(&msgTx.SubnetworkID)
			if err != nil {
				return err
			}
			if gasUsageInSubnetwork > gasLimit {
				str := fmt.Sprintf("Block wastes too much gas in subnetwork with ID %s", msgTx.SubnetworkID)
				return ruleError(ErrInvalidGas, str)
			}
		}
	}
	return nil
}

// LastFinalityPointHash returns the hash of the last finality point
func (dag *BlockDAG) LastFinalityPointHash() *daghash.Hash {
	if dag.lastFinalityPoint == nil {
		return nil
	}
	return &dag.lastFinalityPoint.hash
}

// checkFinalityRules checks the new block does not violate the finality rules
// specifically - the new block selectedParent chain should contain the old finality point
func (dag *BlockDAG) checkFinalityRules(newNode *blockNode) error {
	// the genesis block can not violate finality rules
	if newNode.isGenesis() {
		return nil
	}

	for currentNode := newNode; currentNode != dag.lastFinalityPoint; currentNode = currentNode.selectedParent {
		// If we went past dag's last finality point without encountering it -
		// the new block has violated finality.
		if currentNode.blueScore <= dag.lastFinalityPoint.blueScore {
			return ruleError(ErrFinality, "The last finality point is not in the selected chain of this block")
		}
	}
	return nil
}

// newFinalityPoint return the (potentially) new finality point after the the introduction of a new block
func (dag *BlockDAG) newFinalityPoint(newNode *blockNode) *blockNode {
	// if the new node is the genesis block - it should be the new finality point
	if newNode.isGenesis() {
		return newNode
	}

	// We are looking for a new finality point only if the new block's finality score is higher
	// than the existing finality point's
	if newNode.finalityScore() <= dag.lastFinalityPoint.finalityScore() {
		return dag.lastFinalityPoint
	}

	var currentNode *blockNode
	for currentNode = newNode.selectedParent; ; currentNode = currentNode.selectedParent {
		// If current node's finality score is higher than it's selectedParent's -
		// current node is the new finalityPoint
		if currentNode.isGenesis() || currentNode.finalityScore() > currentNode.selectedParent.finalityScore() {
			break
		}
	}

	return currentNode
}

// NextBlockFeeTransaction prepares the fee transaction for the next mined block
func (dag *BlockDAG) NextBlockFeeTransaction() (*wire.MsgTx, error) {
	_, txsAcceptanceData, _, err := dag.virtual.blockNode.verifyAndBuildUTXO(dag, nil, true)
	if err != nil {
		return nil, err
	}
	return dag.virtual.blockNode.buildFeeTransaction(dag, txsAcceptanceData)
}

// applyDAGChanges does the following:
// 1. Connects each of the new block's parents to the block.
// 2. Adds the new block to the DAG's tips.
// 3. Updates the DAG's full UTXO set.
// 4. Updates each of the tips' utxoDiff.
// 5. Update the finality point of the DAG (if required).
//
// It returns the diff in the virtual block's UTXO set.
//
// This function MUST be called with the DAG state lock held (for writes).
func (dag *BlockDAG) applyDAGChanges(node *blockNode, block *util.Block, newBlockUTXO UTXOSet, fastAdd bool) (
	virtualUTXODiff *UTXODiff, err error) {

	// Clone the virtual block so that we don't modify the existing one.
	virtualClone := dag.virtual.clone()

	if err = node.updateParents(virtualClone, newBlockUTXO); err != nil {
		return nil, fmt.Errorf("failed updating parents of %s: %s", node, err)
	}

	// Update the virtual block's children (the DAG tips) to include the new block.
	virtualClone.AddTip(node)

	// Build a UTXO set for the new virtual block
	newVirtualUTXO, _, err := virtualClone.blockNode.pastUTXO(virtualClone, dag.db)
	if err != nil {
		return nil, fmt.Errorf("could not restore past UTXO for virtual %s: %s", virtualClone, err)
	}

	// Apply new utxoDiffs to all the tips
	err = updateTipsUTXO(virtualClone.parents, virtualClone, newVirtualUTXO)
	if err != nil {
		return nil, fmt.Errorf("failed updating the tips' UTXO: %s", err)
	}

	// It is now safe to meld the UTXO set to base.
	diffSet := newVirtualUTXO.(*DiffUTXOSet)
	virtualUTXODiff = diffSet.UTXODiff
	dag.meldVirtualUTXO(diffSet)

	// Set the status to valid for all index nodes to make sure the changes get
	// written to the database.
	for _, p := range node.parents {
		dag.index.SetStatusFlags(p, statusValid)
	}
	dag.index.SetStatusFlags(node, statusValid)

	// It is now safe to apply the new virtual block
	dag.virtual = virtualClone

	// And now we can update the finality point of the DAG (if required)
	dag.lastFinalityPoint = dag.newFinalityPoint(node)

	return virtualUTXODiff, nil
}

func (dag *BlockDAG) meldVirtualUTXO(newVirtualUTXODiffSet *DiffUTXOSet) {
	dag.utxoLock.Lock()
	defer dag.utxoLock.Unlock()
	newVirtualUTXODiffSet.meldToBase()
}

func (node *blockNode) diffFromTxs(pastUTXO UTXOSet, transactions []*util.Tx) (*UTXODiff, error) {
	diff := NewUTXODiff()

	for _, tx := range transactions {
		txDiff, err := pastUTXO.diffFromTx(tx.MsgTx(), node)
		if err != nil {
			return nil, err
		}
		diff, err = diff.WithDiff(txDiff)
		if err != nil {
			return nil, err
		}
	}

	return diff, nil
}

func (node *blockNode) addTxsToAcceptanceData(txsAcceptanceData MultiblockTxsAcceptanceData, transactions []*util.Tx) {
	blockTxsAcceptanceData := BlockTxsAcceptanceData{}
	for _, tx := range transactions {
		blockTxsAcceptanceData = append(blockTxsAcceptanceData, TxAcceptanceData{
			Tx:         tx,
			IsAccepted: true,
		})
	}
	txsAcceptanceData[node.hash] = blockTxsAcceptanceData
}

// verifyAndBuildUTXO verifies all transactions in the given block and builds its UTXO
// to save extra traversals it returns the transactions acceptance data and the compactFeeData for the new block
func (node *blockNode) verifyAndBuildUTXO(dag *BlockDAG, transactions []*util.Tx, fastAdd bool) (
	newBlockUTXO UTXOSet, txsAcceptanceData MultiblockTxsAcceptanceData, newBlockFeeData compactFeeData, err error) {

	pastUTXO, txsAcceptanceData, err := node.pastUTXO(dag.virtual, dag.db)
	if err != nil {
		return nil, nil, nil, err
	}

	feeData, err := dag.checkConnectToPastUTXO(node, pastUTXO, transactions, fastAdd)
	if err != nil {
		return nil, nil, nil, err
	}

	diff, err := node.diffFromTxs(pastUTXO, transactions)

	node.addTxsToAcceptanceData(txsAcceptanceData, transactions)

	utxo, err := pastUTXO.WithDiff(diff)
	if err != nil {
		return nil, nil, nil, err
	}
	return utxo, txsAcceptanceData, feeData, nil
}

// TxAcceptanceData stores a transaction together with an indication
// if it was accepted or not by some block
type TxAcceptanceData struct {
	Tx         *util.Tx
	IsAccepted bool
}

// BlockTxsAcceptanceData  stores all transactions in a block with an indication
// if they were accepted or not by some other block
type BlockTxsAcceptanceData []TxAcceptanceData

// MultiblockTxsAcceptanceData  stores data about which transactions were accepted by a block
// It's a map from the block's blues block IDs to the transaction acceptance data
type MultiblockTxsAcceptanceData map[daghash.Hash]BlockTxsAcceptanceData

func genesisPastUTXO(virtual *virtualBlock) UTXOSet {
	// The genesis has no past UTXO, so we create an empty UTXO
	// set by creating a diff UTXO set with the virtual UTXO
	// set, and adding all of its entries in toRemove
	diff := NewUTXODiff()
	for outPoint, entry := range virtual.utxoSet.utxoCollection {
		diff.toRemove[outPoint] = entry
	}
	genesisPastUTXO := UTXOSet(NewDiffUTXOSet(virtual.utxoSet, diff))
	return genesisPastUTXO
}

func (node *blockNode) fetchBlueBlocks(db database.DB) ([]*util.Block, error) {
	// Fetch from the database all the transactions for this block's blue set
	blueBlocks := make([]*util.Block, 0, len(node.blues))
	err := db.View(func(dbTx database.Tx) error {
		// Precalculate the amount of transactions in this block's blue set, besides the selected parent.
		// This is to avoid an attack in which an attacker fabricates a block that will deliberately cause
		// a lot of copying, causing a high cost to the whole network.
		for i := len(node.blues) - 1; i >= 0; i-- {
			blueBlockNode := node.blues[i]

			blueBlock, err := dbFetchBlockByNode(dbTx, blueBlockNode)
			if err != nil {
				return err
			}

			blueBlocks = append(blueBlocks, blueBlock)
		}

		return nil
	})
	return blueBlocks, err
}

// applyBlueBlocks adds all transactions in the blue blocks to the selectedParent's UTXO set
// Purposefully ignoring failures - these are just unaccepted transactions
// Writing down which transactions were accepted or not in txsAcceptanceData
func (node *blockNode) applyBlueBlocks(selectedParentUTXO UTXOSet, blueBlocks []*util.Block) (
	pastUTXO UTXOSet, txsAcceptanceData MultiblockTxsAcceptanceData, err error) {

	pastUTXO = selectedParentUTXO
	txsAcceptanceData = MultiblockTxsAcceptanceData{}

	for _, blueBlock := range blueBlocks {
		transactions := blueBlock.Transactions()
		blockTxsAcceptanceData := make(BlockTxsAcceptanceData, len(transactions))
		isSelectedParent := blueBlock.Hash().IsEqual(&node.selectedParent.hash)
		for i, tx := range blueBlock.Transactions() {
			var isAccepted bool
			if isSelectedParent {
				isAccepted = true
			} else {
				isAccepted = pastUTXO.AddTx(tx.MsgTx(), node.height)
			}
			blockTxsAcceptanceData[i] = TxAcceptanceData{Tx: tx, IsAccepted: isAccepted}
		}
		txsAcceptanceData[*blueBlock.Hash()] = blockTxsAcceptanceData
	}

	return pastUTXO, txsAcceptanceData, nil
}

// pastUTXO returns the UTXO of a given block's past
// To save traversals over the blue blocks, it also returns the transaction acceptance data for
// all blue blocks
func (node *blockNode) pastUTXO(virtual *virtualBlock, db database.DB) (
	pastUTXO UTXOSet, bluesTxsAcceptanceData MultiblockTxsAcceptanceData, err error) {

	if node.isGenesis() {
		return genesisPastUTXO(virtual), MultiblockTxsAcceptanceData{}, nil
	}
	selectedParentUTXO, err := node.selectedParent.restoreUTXO(virtual)
	if err != nil {
		return nil, nil, err
	}

	blueBlocks, err := node.fetchBlueBlocks(db)
	if err != nil {
		return nil, nil, err
	}

	return node.applyBlueBlocks(selectedParentUTXO, blueBlocks)
}

// restoreUTXO restores the UTXO of a given block from its diff
func (node *blockNode) restoreUTXO(virtual *virtualBlock) (UTXOSet, error) {
	stack := []*blockNode{node}
	current := node

	for current.diffChild != nil {
		current = current.diffChild
		stack = append(stack, current)
	}

	utxo := UTXOSet(virtual.utxoSet)

	var err error
	for i := len(stack) - 1; i >= 0; i-- {
		utxo, err = utxo.WithDiff(stack[i].diff)
		if err != nil {
			return nil, err
		}
	}

	return utxo, nil
}

// updateParents adds this block to the children sets of its parents
// and updates the diff of any parent whose DiffChild is this block
func (node *blockNode) updateParents(virtual *virtualBlock, newBlockUTXO UTXOSet) error {
	virtualDiffFromNewBlock, err := virtual.utxoSet.diffFrom(newBlockUTXO)
	if err != nil {
		return err
	}

	node.diff = virtualDiffFromNewBlock

	for _, parent := range node.parents {
		if parent.diffChild == nil {
			parentUTXO, err := parent.restoreUTXO(virtual)
			if err != nil {
				return err
			}
			parent.diffChild = node
			parent.diff, err = newBlockUTXO.diffFrom(parentUTXO)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// updateTipsUTXO builds and applies new diff UTXOs for all the DAG's tips
func updateTipsUTXO(tips blockSet, virtual *virtualBlock, virtualUTXO UTXOSet) error {
	for _, tip := range tips {
		tipUTXO, err := tip.restoreUTXO(virtual)
		if err != nil {
			return err
		}
		tip.diff, err = virtualUTXO.diffFrom(tipUTXO)
		if err != nil {
			return err
		}
	}

	return nil
}

// isCurrent returns whether or not the DAG believes it is current.  Several
// factors are used to guess, but the key factors that allow the DAG to
// believe it is current are:
//  - Latest block height is after the latest checkpoint (if enabled)
//  - Latest block has a timestamp newer than 24 hours ago
//
// This function MUST be called with the DAG state lock held (for reads).
func (dag *BlockDAG) isCurrent() bool {
	// Not current if the virtual's selected tip height is less than
	// the latest known good checkpoint (when checkpoints are enabled).
	checkpoint := dag.LatestCheckpoint()
	if checkpoint != nil && dag.selectedTip().height < checkpoint.Height {
		return false
	}

	// Not current if the virtual's selected parent has a timestamp
	// before 24 hours ago. If the DAG is empty, we take the genesis
	// block timestamp.
	//
	// The DAG appears to be current if none of the checks reported
	// otherwise.
	var dagTimestamp int64
	selectedTip := dag.selectedTip()
	if selectedTip == nil {
		dagTimestamp = dag.dagParams.GenesisBlock.Header.Timestamp.Unix()
	} else {
		dagTimestamp = selectedTip.timestamp
	}
	minus24Hours := dag.timeSource.AdjustedTime().Add(-24 * time.Hour).Unix()
	return dagTimestamp >= minus24Hours
}

// IsCurrent returns whether or not the chain believes it is current.  Several
// factors are used to guess, but the key factors that allow the chain to
// believe it is current are:
//  - Latest block height is after the latest checkpoint (if enabled)
//  - Latest block has a timestamp newer than 24 hours ago
//
// This function is safe for concurrent access.
func (dag *BlockDAG) IsCurrent() bool {
	dag.dagLock.RLock()
	defer dag.dagLock.RUnlock()

	return dag.isCurrent()
}

// selectedTip returns the current selected tip for the DAG.
// It will return nil if there is no tip.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) selectedTip() *blockNode {
	return dag.virtual.selectedParent
}

// SelectedTipHeader returns the header of the current selected tip for the DAG.
// It will return nil if there is no tip.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) SelectedTipHeader() *wire.BlockHeader {
	selectedTip := dag.selectedTip()
	if selectedTip == nil {
		return nil
	}

	return selectedTip.Header()
}

// UTXOSet returns the DAG's UTXO set
func (dag *BlockDAG) UTXOSet() *FullUTXOSet {
	return dag.virtual.utxoSet
}

// CalcPastMedianTime returns the past median time of the DAG.
func (dag *BlockDAG) CalcPastMedianTime() time.Time {
	return dag.virtual.tips().bluest().PastMedianTime()
}

// GetUTXOEntry returns the requested unspent transaction output. The returned
// instance must be treated as immutable since it is shared by all callers.
//
// This function is safe for concurrent access. However, the returned entry (if
// any) is NOT.
func (dag *BlockDAG) GetUTXOEntry(outPoint wire.OutPoint) (*UTXOEntry, bool) {
	return dag.virtual.utxoSet.get(outPoint)
}

// IsInSelectedPathChain returns whether or not a block hash is found in the selected path
func (dag *BlockDAG) IsInSelectedPathChain(blockHash *daghash.Hash) bool {
	return dag.virtual.selectedPathChainSet.containsHash(blockHash)
}

// Height returns the height of the highest tip in the DAG
func (dag *BlockDAG) Height() int32 {
	return dag.virtual.tips().maxHeight()
}

// BlockCount returns the number of blocks in the DAG
func (dag *BlockDAG) BlockCount() uint64 {
	return dag.blockCount
}

// TipHashes returns the hashes of the DAG's tips
func (dag *BlockDAG) TipHashes() []daghash.Hash {
	return dag.virtual.tips().hashes()
}

// HighestTipHash returns the hash of the highest tip.
// This function is a placeholder for places that aren't DAG-compatible, and it's needed to be removed in the future
func (dag *BlockDAG) HighestTipHash() daghash.Hash {
	return dag.virtual.tips().highest().hash
}

// CurrentBits returns the bits of the tip with the lowest bits, which also means it has highest difficulty.
func (dag *BlockDAG) CurrentBits() uint32 {
	tips := dag.virtual.tips()
	minBits := uint32(math.MaxUint32)
	for _, tip := range tips {
		if minBits > tip.Header().Bits {
			minBits = tip.Header().Bits
		}
	}
	return minBits
}

// HeaderByHash returns the block header identified by the given hash or an
// error if it doesn't exist.
func (dag *BlockDAG) HeaderByHash(hash *daghash.Hash) (*wire.BlockHeader, error) {
	node := dag.index.LookupNode(hash)
	if node == nil {
		err := fmt.Errorf("block %s is not known", hash)
		return &wire.BlockHeader{}, err
	}

	return node.Header(), nil
}

// BlockLocatorFromHash returns a block locator for the passed block hash.
// See BlockLocator for details on the algorithm used to create a block locator.
//
// In addition to the general algorithm referenced above, this function will
// return the block locator for the latest known tip of the main (best) chain if
// the passed hash is not currently known.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) BlockLocatorFromHash(hash *daghash.Hash) BlockLocator {
	dag.dagLock.RLock()
	node := dag.index.LookupNode(hash)
	locator := dag.blockLocator(node)
	dag.dagLock.RUnlock()
	return locator
}

// LatestBlockLocator returns a block locator for the latest known tip of the
// main (best) chain.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) LatestBlockLocator() (BlockLocator, error) {
	dag.dagLock.RLock()
	locator := dag.blockLocator(nil)
	dag.dagLock.RUnlock()
	return locator, nil
}

// blockLocator returns a block locator for the passed block node.  The passed
// node can be nil in which case the block locator for the current tip
// associated with the view will be returned.
//
// See the BlockLocator type comments for more details.
//
// This function MUST be called with the DAG state lock held (for reads).
func (dag *BlockDAG) blockLocator(node *blockNode) BlockLocator {
	// Use the selected tip if requested.
	if node == nil {
		node = dag.virtual.selectedParent
	}
	if node == nil {
		return nil
	}

	// Calculate the max number of entries that will ultimately be in the
	// block locator.  See the description of the algorithm for how these
	// numbers are derived.

	// Requested hash itself + genesis block.
	// Then floor(log2(height-10)) entries for the skip portion.
	maxEntries := 2 + util.FastLog2Floor(uint32(node.height))
	locator := make(BlockLocator, 0, maxEntries)

	step := int32(1)
	for node != nil {
		locator = append(locator, &node.hash)

		// Nothing more to add once the genesis block has been added.
		if node.height == 0 {
			break
		}

		// Calculate height of previous node to include ensuring the
		// final node is the genesis block.
		height := node.height - step
		if height < 0 {
			height = 0
		}

		// walk backwards through the nodes to the correct ancestor.
		node = node.SelectedAncestor(height)

		// Double the distance between included hashes.
		step *= 2
	}

	return locator
}

// BlockHeightByHash returns the height of the block with the given hash in the
// DAG.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) BlockHeightByHash(hash *daghash.Hash) (int32, error) {
	node := dag.index.LookupNode(hash)
	if node == nil {
		str := fmt.Sprintf("block %s is not in the DAG", hash)
		return 0, errNotInDAG(str)
	}

	return node.height, nil
}

// ChildHashesByHash returns the child hashes of the block with the given hash in the
// DAG.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) ChildHashesByHash(hash *daghash.Hash) ([]daghash.Hash, error) {
	node := dag.index.LookupNode(hash)
	if node == nil {
		str := fmt.Sprintf("block %s is not in the DAG", hash)
		return nil, errNotInDAG(str)

	}

	return node.children.hashes(), nil
}

// HeightToHashRange returns a range of block hashes for the given start height
// and end hash, inclusive on both ends.  The hashes are for all blocks that are
// ancestors of endHash with height greater than or equal to startHeight.  The
// end hash must belong to a block that is known to be valid.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) HeightToHashRange(startHeight int32,
	endHash *daghash.Hash, maxResults int) ([]daghash.Hash, error) {

	endNode := dag.index.LookupNode(endHash)
	if endNode == nil {
		return nil, fmt.Errorf("no known block header with hash %s", endHash)
	}
	if !dag.index.NodeStatus(endNode).KnownValid() {
		return nil, fmt.Errorf("block %s is not yet validated", endHash)
	}
	endHeight := endNode.height

	if startHeight < 0 {
		return nil, fmt.Errorf("start height (%d) is below 0", startHeight)
	}
	if startHeight > endHeight {
		return nil, fmt.Errorf("start height (%d) is past end height (%d)",
			startHeight, endHeight)
	}

	resultsLength := int(endHeight - startHeight + 1)
	if resultsLength > maxResults {
		return nil, fmt.Errorf("number of results (%d) would exceed max (%d)",
			resultsLength, maxResults)
	}

	// Walk backwards from endHeight to startHeight, collecting block hashes.
	node := endNode
	hashes := make([]daghash.Hash, resultsLength)
	for i := resultsLength - 1; i >= 0; i-- {
		hashes[i] = node.hash
		node = node.selectedParent
	}
	return hashes, nil
}

// IntervalBlockHashes returns hashes for all blocks that are ancestors of
// endHash where the block height is a positive multiple of interval.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) IntervalBlockHashes(endHash *daghash.Hash, interval int,
) ([]daghash.Hash, error) {

	endNode := dag.index.LookupNode(endHash)
	if endNode == nil {
		return nil, fmt.Errorf("no known block header with hash %s", endHash)
	}
	if !dag.index.NodeStatus(endNode).KnownValid() {
		return nil, fmt.Errorf("block %s is not yet validated", endHash)
	}
	endHeight := endNode.height

	resultsLength := int(endHeight) / interval
	hashes := make([]daghash.Hash, resultsLength)

	dag.virtual.mtx.Lock()
	defer dag.virtual.mtx.Unlock()

	blockNode := endNode
	for index := int(endHeight) / interval; index > 0; index-- {
		blockHeight := int32(index * interval)
		blockNode = blockNode.SelectedAncestor(blockHeight)

		hashes[index-1] = blockNode.hash
	}

	return hashes, nil
}

// locateInventory returns the node of the block after the first known block in
// the locator along with the number of subsequent nodes needed to either reach
// the provided stop hash or the provided max number of entries.
//
// In addition, there are two special cases:
//
// - When no locators are provided, the stop hash is treated as a request for
//   that block, so it will either return the node associated with the stop hash
//   if it is known, or nil if it is unknown
// - When locators are provided, but none of them are known, nodes starting
//   after the genesis block will be returned
//
// This is primarily a helper function for the locateBlocks and locateHeaders
// functions.
//
// This function MUST be called with the DAG state lock held (for reads).
func (dag *BlockDAG) locateInventory(locator BlockLocator, hashStop *daghash.Hash, maxEntries uint32) (*blockNode, uint32) {
	// There are no block locators so a specific block is being requested
	// as identified by the stop hash.
	stopNode := dag.index.LookupNode(hashStop)
	if len(locator) == 0 {
		if stopNode == nil {
			// No blocks with the stop hash were found so there is
			// nothing to do.
			return nil, 0
		}
		return stopNode, 1
	}

	// Find the most recent locator block hash in the main chain.  In the
	// case none of the hashes in the locator are in the main chain, fall
	// back to the genesis block.
	startNode := dag.genesis
	for _, hash := range locator {
		node := dag.index.LookupNode(hash)
		if node != nil {
			startNode = node
			break
		}
	}

	// Start at the block after the most recently known block.  When there
	// is no next block it means the most recently known block is the tip of
	// the best chain, so there is nothing more to do.
	startNode = startNode.diffChild
	if startNode == nil {
		return nil, 0
	}

	// Calculate how many entries are needed.
	total := uint32((dag.selectedTip().height - startNode.height) + 1)
	if stopNode != nil && stopNode.height >= startNode.height {
		total = uint32((stopNode.height - startNode.height) + 1)
	}
	if total > maxEntries {
		total = maxEntries
	}

	return startNode, total
}

// locateBlocks returns the hashes of the blocks after the first known block in
// the locator until the provided stop hash is reached, or up to the provided
// max number of block hashes.
//
// See the comment on the exported function for more details on special cases.
//
// This function MUST be called with the DAG state lock held (for reads).
func (dag *BlockDAG) locateBlocks(locator BlockLocator, hashStop *daghash.Hash, maxHashes uint32) []daghash.Hash {
	// Find the node after the first known block in the locator and the
	// total number of nodes after it needed while respecting the stop hash
	// and max entries.
	node, total := dag.locateInventory(locator, hashStop, maxHashes)
	if total == 0 {
		return nil
	}

	// Populate and return the found hashes.
	hashes := make([]daghash.Hash, 0, total)
	for i := uint32(0); i < total; i++ {
		hashes = append(hashes, node.hash)
		node = node.diffChild
	}
	return hashes
}

// LocateBlocks returns the hashes of the blocks after the first known block in
// the locator until the provided stop hash is reached, or up to the provided
// max number of block hashes.
//
// In addition, there are two special cases:
//
// - When no locators are provided, the stop hash is treated as a request for
//   that block, so it will either return the stop hash itself if it is known,
//   or nil if it is unknown
// - When locators are provided, but none of them are known, hashes starting
//   after the genesis block will be returned
//
// This function is safe for concurrent access.
func (dag *BlockDAG) LocateBlocks(locator BlockLocator, hashStop *daghash.Hash, maxHashes uint32) []daghash.Hash {
	dag.dagLock.RLock()
	hashes := dag.locateBlocks(locator, hashStop, maxHashes)
	dag.dagLock.RUnlock()
	return hashes
}

// locateHeaders returns the headers of the blocks after the first known block
// in the locator until the provided stop hash is reached, or up to the provided
// max number of block headers.
//
// See the comment on the exported function for more details on special cases.
//
// This function MUST be called with the DAG state lock held (for reads).
func (dag *BlockDAG) locateHeaders(locator BlockLocator, hashStop *daghash.Hash, maxHeaders uint32) []*wire.BlockHeader {
	// Find the node after the first known block in the locator and the
	// total number of nodes after it needed while respecting the stop hash
	// and max entries.
	node, total := dag.locateInventory(locator, hashStop, maxHeaders)
	if total == 0 {
		return nil
	}

	// Populate and return the found headers.
	headers := make([]*wire.BlockHeader, 0, total)
	for i := uint32(0); i < total; i++ {
		headers = append(headers, node.Header())
		node = node.diffChild
	}
	return headers
}

// UTXORLock locks the DAG's UTXO set for reading.
func (dag *BlockDAG) UTXORLock() {
	dag.dagLock.RLock()
}

// UTXORUnlock unlocks the DAG's UTXO set for reading.
func (dag *BlockDAG) UTXORUnlock() {
	dag.dagLock.RUnlock()
}

// LocateHeaders returns the headers of the blocks after the first known block
// in the locator until the provided stop hash is reached, or up to a max of
// wire.MaxBlockHeadersPerMsg headers.
//
// In addition, there are two special cases:
//
// - When no locators are provided, the stop hash is treated as a request for
//   that header, so it will either return the header for the stop hash itself
//   if it is known, or nil if it is unknown
// - When locators are provided, but none of them are known, headers starting
//   after the genesis block will be returned
//
// This function is safe for concurrent access.
func (dag *BlockDAG) LocateHeaders(locator BlockLocator, hashStop *daghash.Hash) []*wire.BlockHeader {
	dag.dagLock.RLock()
	headers := dag.locateHeaders(locator, hashStop, wire.MaxBlockHeadersPerMsg)
	dag.dagLock.RUnlock()
	return headers
}

// SubnetworkID returns the node's subnetwork ID
func (dag *BlockDAG) SubnetworkID() *subnetworkid.SubnetworkID {
	return dag.subnetworkID
}

// IndexManager provides a generic interface that is called when blocks are
// connected and disconnected to and from the tip of the main chain for the
// purpose of supporting optional indexes.
type IndexManager interface {
	// Init is invoked during chain initialize in order to allow the index
	// manager to initialize itself and any indexes it is managing.  The
	// channel parameter specifies a channel the caller can close to signal
	// that the process should be interrupted.  It can be nil if that
	// behavior is not desired.
	Init(database.DB, *BlockDAG, <-chan struct{}) error

	// ConnectBlock is invoked when a new block has been connected to the
	// DAG.
	ConnectBlock(database.Tx, *util.Block, *BlockDAG, MultiblockTxsAcceptanceData) error
}

// Config is a descriptor which specifies the blockchain instance configuration.
type Config struct {
	// DB defines the database which houses the blocks and will be used to
	// store all metadata created by this package such as the utxo set.
	//
	// This field is required.
	DB database.DB

	// Interrupt specifies a channel the caller can close to signal that
	// long running operations, such as catching up indexes or performing
	// database migrations, should be interrupted.
	//
	// This field can be nil if the caller does not desire the behavior.
	Interrupt <-chan struct{}

	// DAGParams identifies which DAG parameters the DAG is associated
	// with.
	//
	// This field is required.
	DAGParams *dagconfig.Params

	// Checkpoints hold caller-defined checkpoints that should be added to
	// the default checkpoints in DAGParams.  Checkpoints must be sorted
	// by height.
	//
	// This field can be nil if the caller does not wish to specify any
	// checkpoints.
	Checkpoints []dagconfig.Checkpoint

	// TimeSource defines the median time source to use for things such as
	// block processing and determining whether or not the chain is current.
	//
	// The caller is expected to keep a reference to the time source as well
	// and add time samples from other peers on the network so the local
	// time is adjusted to be in agreement with other peers.
	TimeSource MedianTimeSource

	// SigCache defines a signature cache to use when when validating
	// signatures.  This is typically most useful when individual
	// transactions are already being validated prior to their inclusion in
	// a block such as what is usually done via a transaction memory pool.
	//
	// This field can be nil if the caller is not interested in using a
	// signature cache.
	SigCache *txscript.SigCache

	// IndexManager defines an index manager to use when initializing the
	// chain and connecting and disconnecting blocks.
	//
	// This field can be nil if the caller does not wish to make use of an
	// index manager.
	IndexManager IndexManager

	// SubnetworkID identifies which subnetwork the DAG is associated
	// with.
	//
	// This field is required.
	SubnetworkID *subnetworkid.SubnetworkID
}

// New returns a BlockDAG instance using the provided configuration details.
func New(config *Config) (*BlockDAG, error) {
	// Enforce required config fields.
	if config.DB == nil {
		return nil, AssertError("BlockDAG.New database is nil")
	}
	if config.DAGParams == nil {
		return nil, AssertError("BlockDAG.New DAG parameters nil")
	}
	if config.TimeSource == nil {
		return nil, AssertError("BlockDAG.New timesource is nil")
	}
	if config.SubnetworkID == nil {
		return nil, AssertError("BlockDAG.New subnetworkID is nil")
	}

	// Generate a checkpoint by height map from the provided checkpoints
	// and assert the provided checkpoints are sorted by height as required.
	var checkpointsByHeight map[int32]*dagconfig.Checkpoint
	var prevCheckpointHeight int32
	if len(config.Checkpoints) > 0 {
		checkpointsByHeight = make(map[int32]*dagconfig.Checkpoint)
		for i := range config.Checkpoints {
			checkpoint := &config.Checkpoints[i]
			if checkpoint.Height <= prevCheckpointHeight {
				return nil, AssertError("blockdag.New " +
					"checkpoints are not sorted by height")
			}

			checkpointsByHeight[checkpoint.Height] = checkpoint
			prevCheckpointHeight = checkpoint.Height
		}
	}

	params := config.DAGParams
	targetTimespan := int64(params.TargetTimespan / time.Second)
	targetTimePerBlock := int64(params.TargetTimePerBlock / time.Second)
	adjustmentFactor := params.RetargetAdjustmentFactor
	index := newBlockIndex(config.DB, params)
	dag := BlockDAG{
		checkpoints:         config.Checkpoints,
		checkpointsByHeight: checkpointsByHeight,
		db:                  config.DB,
		dagParams:           params,
		timeSource:          config.TimeSource,
		sigCache:            config.SigCache,
		indexManager:        config.IndexManager,
		minRetargetTimespan: targetTimespan / adjustmentFactor,
		maxRetargetTimespan: targetTimespan * adjustmentFactor,
		blocksPerRetarget:   int32(targetTimespan / targetTimePerBlock),
		index:               index,
		virtual:             newVirtualBlock(nil, params.K),
		orphans:             make(map[daghash.Hash]*orphanBlock),
		prevOrphans:         make(map[daghash.Hash][]*orphanBlock),
		warningCaches:       newThresholdCaches(vbNumBits),
		deploymentCaches:    newThresholdCaches(dagconfig.DefinedDeployments),
		blockCount:          0,
		SubnetworkStore:     newSubnetworkStore(config.DB),
		subnetworkID:        config.SubnetworkID,
	}

	// Initialize the chain state from the passed database.  When the db
	// does not yet contain any DAG state, both it and the DAG state
	// will be initialized to contain only the genesis block.
	if err := dag.initDAGState(); err != nil {
		return nil, err
	}

	// Initialize and catch up all of the currently active optional indexes
	// as needed.
	if config.IndexManager != nil {
		err := config.IndexManager.Init(dag.db, &dag, config.Interrupt)
		if err != nil {
			return nil, err
		}
	}

	genesis := index.LookupNode(params.GenesisHash)

	if genesis == nil {
		genesisBlock := util.NewBlock(dag.dagParams.GenesisBlock)
		isOrphan, err := dag.ProcessBlock(genesisBlock, BFNone)
		if err != nil {
			return nil, err
		}
		if isOrphan {
			return nil, errors.New("Genesis block is unexpectedly orphan")
		}
		genesis = index.LookupNode(params.GenesisHash)
	}

	// Save a reference to the genesis block.
	dag.genesis = genesis

	// Initialize rule change threshold state caches.
	if err := dag.initThresholdCaches(); err != nil {
		return nil, err
	}

	selectedTip := dag.selectedTip()
	log.Infof("DAG state (height %d, hash %s, work %d)",
		selectedTip.height, selectedTip.hash, selectedTip.workSum)

	return &dag, nil
}
