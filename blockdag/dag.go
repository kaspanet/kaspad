// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/util/subnetworkid"

	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/database"
	"github.com/kaspanet/kaspad/txscript"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

const (
	// maxOrphanBlocks is the maximum number of orphan blocks that can be
	// queued.
	maxOrphanBlocks = 100

	isDAGCurrentMaxDiff = 12 * time.Hour
)

// orphanBlock represents a block that we don't yet have the parent for. It
// is a normal block plus an expiration time to prevent caching the orphan
// forever.
type orphanBlock struct {
	block      *util.Block
	expiration time.Time
}

// delayedBlock represents a block which has a delayed timestamp and will be processed at processTime
type delayedBlock struct {
	block       *util.Block
	processTime time.Time
}

// chainUpdates represents the updates made to the selected parent chain after
// a block had been added to the DAG.
type chainUpdates struct {
	removedChainBlockHashes []*daghash.Hash
	addedChainBlockHashes   []*daghash.Hash
}

// BlockDAG provides functions for working with the kaspa block DAG.
// It includes functionality such as rejecting duplicate blocks, ensuring blocks
// follow all rules, and orphan handling.
type BlockDAG struct {
	// The following fields are set when the instance is created and can't
	// be changed afterwards, so there is no need to protect them with a
	// separate mutex.
	db           database.DB
	dagParams    *dagconfig.Params
	timeSource   TimeSource
	sigCache     *txscript.SigCache
	indexManager IndexManager
	genesis      *blockNode

	// The following fields are calculated based upon the provided DAG
	// parameters. They are also set when the instance is created and
	// can't be changed afterwards, so there is no need to protect them with
	// a separate mutex.
	targetTimePerBlock             int64 // The target delay between blocks (in seconds)
	difficultyAdjustmentWindowSize uint64
	TimestampDeviationTolerance    uint64

	// powMaxBits defines the highest allowed proof of work value for a
	// block in compact form.
	powMaxBits uint32

	// dagLock protects concurrent access to the vast majority of the
	// fields in this struct below this point.
	dagLock sync.RWMutex

	utxoLock sync.RWMutex

	// index and virtual are related to the memory block index. They both
	// have their own locks, however they are often also protected by the
	// DAG lock to help prevent logic races when blocks are being processed.

	// index houses the entire block index in memory. The block index is
	// a tree-shaped structure.
	index *blockIndex

	// blockCount holds the number of blocks in the DAG
	blockCount uint64

	// virtual tracks the current tips.
	virtual *virtualBlock

	// subnetworkID holds the subnetwork ID of the DAG
	subnetworkID *subnetworkid.SubnetworkID

	// These fields are related to handling of orphan blocks. They are
	// protected by a combination of the DAG lock and the orphan lock.
	orphanLock   sync.RWMutex
	orphans      map[daghash.Hash]*orphanBlock
	prevOrphans  map[daghash.Hash][]*orphanBlock
	newestOrphan *orphanBlock

	// delayedBlocks is a list of all delayed blocks. We are maintaining this
	// list for the case where a new block with a valid timestamp points to a delayed block.
	// In that case we will delay the processing of the child block so it would be processed
	// after its parent.
	delayedBlocks      map[daghash.Hash]*delayedBlock
	delayedBlocksQueue delayedBlocksHeap

	// The following caches are used to efficiently keep track of the
	// current deployment threshold state of each rule change deployment.
	//
	// This information is stored in the database so it can be quickly
	// reconstructed on load.
	//
	// warningCaches caches the current deployment threshold state for blocks
	// in each of the **possible** deployments. This is used in order to
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
	// certain blockDAG events.
	notificationsLock sync.RWMutex
	notifications     []NotificationCallback

	lastFinalityPoint *blockNode

	SubnetworkStore   *SubnetworkStore
	utxoDiffStore     *utxoDiffStore
	reachabilityStore *reachabilityStore
}

// IsKnownBlock returns whether or not the DAG instance has the block represented
// by the passed hash. This includes checking the various places a block can
// be in, like part of the DAG or the orphan pool.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) IsKnownBlock(hash *daghash.Hash) bool {
	return dag.IsInDAG(hash) || dag.IsKnownOrphan(hash) || dag.isKnownDelayedBlock(hash)
}

// AreKnownBlocks returns whether or not the DAG instances has all blocks represented
// by the passed hashes. This includes checking the various places a block can
// be in, like part of the DAG or the orphan pool.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) AreKnownBlocks(hashes []*daghash.Hash) bool {
	for _, hash := range hashes {
		haveBlock := dag.IsKnownBlock(hash)
		if !haveBlock {
			return false
		}
	}

	return true
}

// IsKnownOrphan returns whether the passed hash is currently a known orphan.
// Keep in mind that only a limited number of orphans are held onto for a
// limited amount of time, so this function must not be used as an absolute
// way to test if a block is an orphan block. A full block (as opposed to just
// its hash) must be passed to ProcessBlock for that purpose. However, calling
// ProcessBlock with an orphan that already exists results in an error, so this
// function provides a mechanism for a caller to intelligently detect *recent*
// duplicate orphans and react accordingly.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) IsKnownOrphan(hash *daghash.Hash) bool {
	// Protect concurrent access. Using a read lock only so multiple
	// readers can query without blocking each other.
	dag.orphanLock.RLock()
	defer dag.orphanLock.RUnlock()
	_, exists := dag.orphans[*hash]

	return exists
}

// IsKnownInvalid returns whether the passed hash is known to be an invalid block.
// Note that if the block is not found this method will return false.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) IsKnownInvalid(hash *daghash.Hash) bool {
	node := dag.index.LookupNode(hash)
	if node == nil {
		return false
	}
	return dag.index.NodeStatus(node).KnownInvalid()
}

// GetOrphanMissingAncestorHashes returns all of the missing parents in the orphan's sub-DAG
//
// This function is safe for concurrent access.
func (dag *BlockDAG) GetOrphanMissingAncestorHashes(orphanHash *daghash.Hash) ([]*daghash.Hash, error) {
	// Protect concurrent access. Using a read lock only so multiple
	// readers can query without blocking each other.
	dag.orphanLock.RLock()
	defer dag.orphanLock.RUnlock()

	missingAncestorsHashes := make([]*daghash.Hash, 0)

	visited := make(map[daghash.Hash]bool)
	queue := []*daghash.Hash{orphanHash}
	for len(queue) > 0 {
		var current *daghash.Hash
		current, queue = queue[0], queue[1:]
		if !visited[*current] {
			visited[*current] = true
			orphan, orphanExists := dag.orphans[*current]
			if orphanExists {
				queue = append(queue, orphan.block.MsgBlock().Header.ParentHashes...)
			} else {
				if !dag.IsInDAG(current) && current != orphanHash {
					missingAncestorsHashes = append(missingAncestorsHashes, current)
				}
			}
		}
	}
	return missingAncestorsHashes, nil
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

	// Remove the reference from the previous orphan index too.
	for _, parentHash := range orphan.block.MsgBlock().Header.ParentHashes {
		// An indexing for loop is intentionally used over a range here as range
		// does not reevaluate the slice on each iteration nor does it adjust the
		// index for the modified slice.
		orphans := dag.prevOrphans[*parentHash]
		for i := 0; i < len(orphans); i++ {
			hash := orphans[i].block.Hash()
			if hash.IsEqual(orphanHash) {
				orphans = append(orphans[:i], orphans[i+1:]...)
				i--
			}
		}

		// Remove the map entry altogether if there are no longer any orphans
		// which depend on the parent hash.
		if len(orphans) == 0 {
			delete(dag.prevOrphans, *parentHash)
			continue
		}

		dag.prevOrphans[*parentHash] = orphans
	}
}

// addOrphanBlock adds the passed block (which is already determined to be
// an orphan prior calling this function) to the orphan pool. It lazily cleans
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

		// Update the newest orphan block pointer so it can be discarded
		// in case the orphan pool fills up.
		if dag.newestOrphan == nil || oBlock.block.Timestamp().After(dag.newestOrphan.block.Timestamp()) {
			dag.newestOrphan = oBlock
		}
	}

	// Limit orphan blocks to prevent memory exhaustion.
	if len(dag.orphans)+1 > maxOrphanBlocks {
		// If the new orphan is newer than the newest orphan on the orphan
		// pool, don't add it.
		if block.Timestamp().After(dag.newestOrphan.block.Timestamp()) {
			return
		}
		// Remove the newest orphan to make room for the added one.
		dag.removeOrphanBlock(dag.newestOrphan)
		dag.newestOrphan = nil
	}

	// Protect concurrent access. This is intentionally done here instead
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
	for _, parentHash := range block.MsgBlock().Header.ParentHashes {
		dag.prevOrphans[*parentHash] = append(dag.prevOrphans[*parentHash], oBlock)
	}
}

// SequenceLock represents the converted relative lock-time in seconds, and
// absolute block-blue-score for a transaction input's relative lock-times.
// According to SequenceLock, after the referenced input has been confirmed
// within a block, a transaction spending that input can be included into a
// block either after 'seconds' (according to past median time), or once the
// 'BlockBlueScore' has been reached.
type SequenceLock struct {
	Seconds        int64
	BlockBlueScore int64
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
	dag.dagLock.RLock()
	defer dag.dagLock.RUnlock()

	return dag.calcSequenceLock(dag.selectedTip(), utxoSet, tx, mempool)
}

// CalcSequenceLockNoLock is lock free version of CalcSequenceLockWithLock
// This function is unsafe for concurrent access.
func (dag *BlockDAG) CalcSequenceLockNoLock(tx *util.Tx, utxoSet UTXOSet, mempool bool) (*SequenceLock, error) {
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
	sequenceLock := &SequenceLock{Seconds: -1, BlockBlueScore: -1}

	// Sequence locks don't apply to coinbase transactions Therefore, we
	// return sequence lock values of -1 indicating that this transaction
	// can be included within a block at any given height or time.
	if tx.IsCoinBase() {
		return sequenceLock, nil
	}

	mTx := tx.MsgTx()
	for txInIndex, txIn := range mTx.TxIn {
		entry, ok := utxoSet.Get(txIn.PreviousOutpoint)
		if !ok {
			str := fmt.Sprintf("output %s referenced from "+
				"transaction %s input %d either does not exist or "+
				"has already been spent", txIn.PreviousOutpoint,
				tx.ID(), txInIndex)
			return sequenceLock, ruleError(ErrMissingTxOut, str)
		}

		// If the input blue score is set to the mempool blue score, then we
		// assume the transaction makes it into the next block when
		// evaluating its sequence blocks.
		inputBlueScore := entry.BlockBlueScore()
		if entry.IsUnaccepted() {
			inputBlueScore = dag.virtual.blueScore
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
			// in seconds before it can be spent. Therefore, we
			// need to query for the block prior to the one in
			// which this input was accepted within so we can
			// compute the past median time for the block prior to
			// the one which accepted this referenced output.
			blockNode := node
			for blockNode.selectedParent.blueScore > inputBlueScore {
				blockNode = blockNode.selectedParent
			}
			medianTime := blockNode.PastMedianTime(dag)

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
			// the input's blue score as its converted absolute
			// lock-time. We subtract one from the relative lock in
			// order to maintain the original lockTime semantics.
			blockBlueScore := int64(inputBlueScore) + relativeLock - 1
			if blockBlueScore > sequenceLock.BlockBlueScore {
				sequenceLock.BlockBlueScore = blockBlueScore
			}
		}
	}

	return sequenceLock, nil
}

// LockTimeToSequence converts the passed relative locktime to a sequence
// number in accordance to BIP-68.
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
//
// This function MUST be called with the DAG state lock held (for writes).
func (dag *BlockDAG) addBlock(node *blockNode,
	block *util.Block, selectedParentAnticone []*blockNode, flags BehaviorFlags) (*chainUpdates, error) {
	// Skip checks if node has already been fully validated.
	fastAdd := flags&BFFastAdd == BFFastAdd || dag.index.NodeStatus(node).KnownValid()

	// Connect the block to the DAG.
	chainUpdates, err := dag.connectBlock(node, block, selectedParentAnticone, fastAdd)
	if err != nil {
		if errors.As(err, &RuleError{}) {
			dag.index.SetStatusFlags(node, statusValidateFailed)
			err := dag.index.flushToDB()
			if err != nil {
				return nil, err
			}
		}
		return nil, err
	}
	dag.blockCount++
	return chainUpdates, nil
}

func calculateAcceptedIDMerkleRoot(multiBlockTxsAcceptanceData MultiBlockTxsAcceptanceData) *daghash.Hash {
	var acceptedTxs []*util.Tx
	for _, blockTxsAcceptanceData := range multiBlockTxsAcceptanceData {
		for _, txAcceptance := range blockTxsAcceptanceData.TxAcceptanceData {
			if !txAcceptance.IsAccepted {
				continue
			}
			acceptedTxs = append(acceptedTxs, txAcceptance.Tx)
		}
	}
	sort.Slice(acceptedTxs, func(i, j int) bool {
		return daghash.LessTxID(acceptedTxs[i].ID(), acceptedTxs[j].ID())
	})

	acceptedIDMerkleTree := BuildIDMerkleTreeStore(acceptedTxs)
	return acceptedIDMerkleTree.Root()
}

func (node *blockNode) validateAcceptedIDMerkleRoot(dag *BlockDAG, txsAcceptanceData MultiBlockTxsAcceptanceData) error {
	if node.isGenesis() {
		return nil
	}

	calculatedAccepetedIDMerkleRoot := calculateAcceptedIDMerkleRoot(txsAcceptanceData)
	header := node.Header()
	if !header.AcceptedIDMerkleRoot.IsEqual(calculatedAccepetedIDMerkleRoot) {
		str := fmt.Sprintf("block accepted ID merkle root is invalid - block "+
			"header indicates %s, but calculated value is %s",
			header.AcceptedIDMerkleRoot, calculatedAccepetedIDMerkleRoot)
		return ruleError(ErrBadMerkleRoot, str)
	}
	return nil
}

// connectBlock handles connecting the passed node/block to the DAG.
//
// This function MUST be called with the DAG state lock held (for writes).
func (dag *BlockDAG) connectBlock(node *blockNode,
	block *util.Block, selectedParentAnticone []*blockNode, fastAdd bool) (*chainUpdates, error) {
	// No warnings about unknown rules or versions until the DAG is
	// current.
	if dag.isCurrent() {
		// Warn if any unknown new rules are either about to activate or
		// have already been activated.
		if err := dag.warnUnknownRuleActivations(node); err != nil {
			return nil, err
		}

		// Warn if a high enough percentage of the last blocks have
		// unexpected versions.
		if err := dag.warnUnknownVersions(node); err != nil {
			return nil, err
		}
	}

	if err := dag.checkFinalityRules(node); err != nil {
		return nil, err
	}

	if err := dag.validateGasLimit(block); err != nil {
		return nil, err
	}

	newBlockUTXO, txsAcceptanceData, newBlockFeeData, err := node.verifyAndBuildUTXO(dag, block.Transactions(), fastAdd)
	if err != nil {
		newErrString := fmt.Sprintf("error verifying UTXO for %s: %s", node, err)
		var ruleErr RuleError
		if ok := errors.As(err, &ruleErr); ok {
			return nil, ruleError(ruleErr.ErrorCode, newErrString)
		}
		return nil, errors.New(newErrString)
	}

	err = node.validateCoinbaseTransaction(dag, block, txsAcceptanceData)
	if err != nil {
		return nil, err
	}

	// Apply all changes to the DAG.
	virtualUTXODiff, virtualTxsAcceptanceData, chainUpdates, err := dag.applyDAGChanges(node, newBlockUTXO, selectedParentAnticone)
	if err != nil {
		// Since all validation logic has already ran, if applyDAGChanges errors out,
		// this means we have a problem in the internal structure of the DAG - a problem which is
		// irrecoverable, and it would be a bad idea to attempt adding any more blocks to the DAG.
		// Therefore - in such cases we panic.
		panic(err)
	}

	err = dag.saveChangesFromBlock(block, virtualUTXODiff, txsAcceptanceData, virtualTxsAcceptanceData, newBlockFeeData)
	if err != nil {
		return nil, err
	}

	return chainUpdates, nil
}

func (dag *BlockDAG) saveChangesFromBlock(block *util.Block, virtualUTXODiff *UTXODiff,
	txsAcceptanceData MultiBlockTxsAcceptanceData, virtualTxsAcceptanceData MultiBlockTxsAcceptanceData,
	feeData compactFeeData) error {

	// Atomically insert info into the database.
	err := dag.db.Update(func(dbTx database.Tx) error {
		err := dag.index.flushToDBWithTx(dbTx)
		if err != nil {
			return err
		}

		err = dag.utxoDiffStore.flushToDB(dbTx)
		if err != nil {
			return err
		}

		err = dag.reachabilityStore.flushToDB(dbTx)
		if err != nil {
			return err
		}

		// Update best block state.
		state := &dagState{
			TipHashes:         dag.TipHashes(),
			LastFinalityPoint: dag.lastFinalityPoint.hash,
		}
		err = dbPutDAGState(dbTx, state)
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
		err = registerSubnetworks(dbTx, block.Transactions())
		if err != nil {
			return err
		}

		blockID, err := createBlockID(dbTx, block.Hash())
		if err != nil {
			return err
		}

		// Allow the index manager to call each of the currently active
		// optional indexes with the block being connected so they can
		// update themselves accordingly.
		if dag.indexManager != nil {
			err := dag.indexManager.ConnectBlock(dbTx, block, blockID, dag, txsAcceptanceData, virtualTxsAcceptanceData)
			if err != nil {
				return err
			}
		}

		// Apply the fee data into the database
		return dbStoreFeeData(dbTx, block.Hash(), feeData)
	})
	if err != nil {
		return err
	}
	dag.index.clearDirtyEntries()
	dag.utxoDiffStore.clearDirtyEntries()
	dag.reachabilityStore.clearDirtyEntries()
	return nil
}

func (dag *BlockDAG) validateGasLimit(block *util.Block) error {
	var currentSubnetworkID *subnetworkid.SubnetworkID
	var currentSubnetworkGasLimit uint64
	var currentGasUsage uint64
	var err error

	// We assume here that transactions are ordered by subnetworkID,
	// since it was already validated in checkTransactionSanity
	for _, tx := range block.Transactions() {
		msgTx := tx.MsgTx()

		// In native and Built-In subnetworks all txs must have Gas = 0, and that was already validated in checkTransactionSanity
		// Therefore - no need to check them here.
		if msgTx.SubnetworkID.IsEqual(subnetworkid.SubnetworkIDNative) || msgTx.SubnetworkID.IsBuiltIn() {
			continue
		}

		if !msgTx.SubnetworkID.IsEqual(currentSubnetworkID) {
			currentSubnetworkID = &msgTx.SubnetworkID
			currentGasUsage = 0
			currentSubnetworkGasLimit, err = dag.SubnetworkStore.GasLimit(currentSubnetworkID)
			if err != nil {
				return errors.Errorf("Error getting gas limit for subnetworkID '%s': %s", currentSubnetworkID, err)
			}
		}

		newGasUsage := currentGasUsage + msgTx.Gas
		if newGasUsage < currentGasUsage { // check for overflow
			str := fmt.Sprintf("Block gas usage in subnetwork with ID %s has overflown", currentSubnetworkID)
			return ruleError(ErrInvalidGas, str)
		}
		if newGasUsage > currentSubnetworkGasLimit {
			str := fmt.Sprintf("Block wastes too much gas in subnetwork with ID %s", currentSubnetworkID)
			return ruleError(ErrInvalidGas, str)
		}

		currentGasUsage = newGasUsage
	}

	return nil
}

// LastFinalityPointHash returns the hash of the last finality point
func (dag *BlockDAG) LastFinalityPointHash() *daghash.Hash {
	if dag.lastFinalityPoint == nil {
		return nil
	}
	return dag.lastFinalityPoint.hash
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

// updateFinalityPoint updates the dag's last finality point if necessary.
func (dag *BlockDAG) updateFinalityPoint() {
	selectedTip := dag.selectedTip()
	// if the selected tip is the genesis block - it should be the new finality point
	if selectedTip.isGenesis() {
		dag.lastFinalityPoint = selectedTip
		return
	}
	// We are looking for a new finality point only if the new block's finality score is higher
	// by 2 than the existing finality point's
	if selectedTip.finalityScore(dag) < dag.lastFinalityPoint.finalityScore(dag)+2 {
		return
	}

	var currentNode *blockNode
	for currentNode = selectedTip.selectedParent; ; currentNode = currentNode.selectedParent {
		// We look for the first node in the selected parent chain that has a higher finality score than the last finality point.
		if currentNode.selectedParent.finalityScore(dag) == dag.lastFinalityPoint.finalityScore(dag) {
			break
		}
	}
	dag.lastFinalityPoint = currentNode
	spawn(func() {
		dag.finalizeNodesBelowFinalityPoint(true)
	})
}

func (dag *BlockDAG) finalizeNodesBelowFinalityPoint(deleteDiffData bool) {
	queue := make([]*blockNode, 0, len(dag.lastFinalityPoint.parents))
	for parent := range dag.lastFinalityPoint.parents {
		queue = append(queue, parent)
	}
	var blockHashesToDelete []*daghash.Hash
	if deleteDiffData {
		blockHashesToDelete = make([]*daghash.Hash, 0, dag.dagParams.FinalityInterval)
	}
	for len(queue) > 0 {
		var current *blockNode
		current, queue = queue[0], queue[1:]
		if !current.isFinalized {
			current.isFinalized = true
			if deleteDiffData {
				blockHashesToDelete = append(blockHashesToDelete, current.hash)
			}
			for parent := range current.parents {
				queue = append(queue, parent)
			}
		}
	}
	if deleteDiffData {
		err := dag.db.Update(func(dbTx database.Tx) error {
			return dag.utxoDiffStore.removeBlocksDiffData(dbTx, blockHashesToDelete)
		})
		if err != nil {
			panic(fmt.Sprintf("Error removing diff data from utxoDiffStore: %s", err))
		}
	}
}

// IsKnownFinalizedBlock returns whether the block is below the finality point.
// IsKnownFinalizedBlock might be false-negative because node finality status is
// updated in a separate goroutine. To get a definite answer if a block
// is finalized or not, use dag.checkFinalityRules.
func (dag *BlockDAG) IsKnownFinalizedBlock(blockHash *daghash.Hash) bool {
	node := dag.index.LookupNode(blockHash)
	return node != nil && node.isFinalized
}

// NextBlockCoinbaseTransaction prepares the coinbase transaction for the next mined block
//
// This function CAN'T be called with the DAG lock held.
func (dag *BlockDAG) NextBlockCoinbaseTransaction(scriptPubKey []byte, extraData []byte) (*util.Tx, error) {
	dag.dagLock.RLock()
	defer dag.dagLock.RUnlock()

	return dag.NextBlockCoinbaseTransactionNoLock(scriptPubKey, extraData)
}

// NextBlockCoinbaseTransactionNoLock prepares the coinbase transaction for the next mined block
//
// This function MUST be called with the DAG read-lock held
func (dag *BlockDAG) NextBlockCoinbaseTransactionNoLock(scriptPubKey []byte, extraData []byte) (*util.Tx, error) {
	txsAcceptanceData, err := dag.TxsAcceptedByVirtual()
	if err != nil {
		return nil, err
	}
	return dag.virtual.blockNode.expectedCoinbaseTransaction(dag, txsAcceptanceData, scriptPubKey, extraData)
}

// NextAcceptedIDMerkleRootNoLock prepares the acceptedIDMerkleRoot for the next mined block
//
// This function MUST be called with the DAG read-lock held
func (dag *BlockDAG) NextAcceptedIDMerkleRootNoLock() (*daghash.Hash, error) {
	txsAcceptanceData, err := dag.TxsAcceptedByVirtual()
	if err != nil {
		return nil, err
	}

	return calculateAcceptedIDMerkleRoot(txsAcceptanceData), nil
}

// TxsAcceptedByVirtual retrieves transactions accepted by the current virtual block
//
// This function MUST be called with the DAG read-lock held
func (dag *BlockDAG) TxsAcceptedByVirtual() (MultiBlockTxsAcceptanceData, error) {
	_, txsAcceptanceData, err := dag.pastUTXO(&dag.virtual.blockNode)
	return txsAcceptanceData, err
}

// TxsAcceptedByBlockHash retrieves transactions accepted by the given block
//
// This function MUST be called with the DAG read-lock held
func (dag *BlockDAG) TxsAcceptedByBlockHash(blockHash *daghash.Hash) (MultiBlockTxsAcceptanceData, error) {
	node := dag.index.LookupNode(blockHash)
	if node == nil {
		return nil, errors.Errorf("Couldn't find block %s", blockHash)
	}
	_, txsAcceptanceData, err := dag.pastUTXO(node)
	return txsAcceptanceData, err
}

// applyDAGChanges does the following:
// 1. Connects each of the new block's parents to the block.
// 2. Adds the new block to the DAG's tips.
// 3. Updates the DAG's full UTXO set.
// 4. Updates each of the tips' utxoDiff.
// 5. Applies the new virtual's blue score to all the unaccepted UTXOs
// 6. Adds the block to the reachability structures
// 7. Updates the finality point of the DAG (if required).
//
// It returns the diff in the virtual block's UTXO set.
//
// This function MUST be called with the DAG state lock held (for writes).
func (dag *BlockDAG) applyDAGChanges(node *blockNode, newBlockUTXO UTXOSet, selectedParentAnticone []*blockNode) (
	virtualUTXODiff *UTXODiff, virtualTxsAcceptanceData MultiBlockTxsAcceptanceData,
	chainUpdates *chainUpdates, err error) {

	// Add the block to the reachability structures
	err = dag.updateReachability(node, selectedParentAnticone)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed updating reachability")
	}

	if err = node.updateParents(dag, newBlockUTXO); err != nil {
		return nil, nil, nil, errors.Wrapf(err, "failed updating parents of %s", node)
	}

	// Update the virtual block's parents (the DAG tips) to include the new block.
	chainUpdates = dag.virtual.AddTip(node)

	// Build a UTXO set for the new virtual block
	newVirtualUTXO, virtualTxsAcceptanceData, err := dag.pastUTXO(&dag.virtual.blockNode)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "could not restore past UTXO for virtual")
	}

	// Apply new utxoDiffs to all the tips
	err = updateTipsUTXO(dag, newVirtualUTXO)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed updating the tips' UTXO")
	}

	// It is now safe to meld the UTXO set to base.
	diffSet := newVirtualUTXO.(*DiffUTXOSet)
	virtualUTXODiff = diffSet.UTXODiff
	err = dag.meldVirtualUTXO(diffSet)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed melding the virtual UTXO")
	}

	dag.index.SetStatusFlags(node, statusValid)

	// And now we can update the finality point of the DAG (if required)
	dag.updateFinalityPoint()

	return virtualUTXODiff, virtualTxsAcceptanceData, chainUpdates, nil
}

func (dag *BlockDAG) meldVirtualUTXO(newVirtualUTXODiffSet *DiffUTXOSet) error {
	dag.utxoLock.Lock()
	defer dag.utxoLock.Unlock()
	return newVirtualUTXODiffSet.meldToBase()
}

func (node *blockNode) diffFromTxs(pastUTXO UTXOSet, transactions []*util.Tx) (*UTXODiff, error) {
	diff := NewUTXODiff()

	for _, tx := range transactions {
		txDiff, err := pastUTXO.diffFromTx(tx.MsgTx(), UnacceptedBlueScore)
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

// verifyAndBuildUTXO verifies all transactions in the given block and builds its UTXO
// to save extra traversals it returns the transactions acceptance data and the compactFeeData for the new block
func (node *blockNode) verifyAndBuildUTXO(dag *BlockDAG, transactions []*util.Tx, fastAdd bool) (
	newBlockUTXO UTXOSet, txsAcceptanceData MultiBlockTxsAcceptanceData, newBlockFeeData compactFeeData, err error) {

	pastUTXO, txsAcceptanceData, err := dag.pastUTXO(node)
	if err != nil {
		return nil, nil, nil, err
	}

	err = node.validateAcceptedIDMerkleRoot(dag, txsAcceptanceData)
	if err != nil {
		return nil, nil, nil, err
	}

	feeData, err := dag.checkConnectToPastUTXO(node, pastUTXO, transactions, fastAdd)
	if err != nil {
		return nil, nil, nil, err
	}

	diffFromTxs, err := node.diffFromTxs(pastUTXO, transactions)
	if err != nil {
		return nil, nil, nil, err
	}
	utxo, err := pastUTXO.WithDiff(diffFromTxs)
	if err != nil {
		return nil, nil, nil, err
	}

	calculatedMultisetHash := utxo.Multiset().Hash()
	if !calculatedMultisetHash.IsEqual(node.utxoCommitment) {
		str := fmt.Sprintf("block %s UTXO commitment is invalid - block "+
			"header indicates %s, but calculated value is %s", node.hash,
			node.utxoCommitment, calculatedMultisetHash)
		return nil, nil, nil, ruleError(ErrBadUTXOCommitment, str)
	}
	return utxo, txsAcceptanceData, feeData, nil
}

// TxAcceptanceData stores a transaction together with an indication
// if it was accepted or not by some block
type TxAcceptanceData struct {
	Tx         *util.Tx
	IsAccepted bool
}

// BlockTxsAcceptanceData stores all transactions in a block with an indication
// if they were accepted or not by some other block
type BlockTxsAcceptanceData struct {
	BlockHash        daghash.Hash
	TxAcceptanceData []TxAcceptanceData
}

// MultiBlockTxsAcceptanceData stores data about which transactions were accepted by a block
// It's a slice of the block's blues block IDs and their transaction acceptance data
type MultiBlockTxsAcceptanceData []BlockTxsAcceptanceData

// FindAcceptanceData finds the BlockTxsAcceptanceData that matches blockHash
func (data MultiBlockTxsAcceptanceData) FindAcceptanceData(blockHash *daghash.Hash) (*BlockTxsAcceptanceData, bool) {
	for _, acceptanceData := range data {
		if acceptanceData.BlockHash.IsEqual(blockHash) {
			return &acceptanceData, true
		}
	}
	return nil, false
}

func genesisPastUTXO(virtual *virtualBlock) UTXOSet {
	// The genesis has no past UTXO, so we create an empty UTXO
	// set by creating a diff UTXO set with the virtual UTXO
	// set, and adding all of its entries in toRemove
	diff := NewUTXODiff()
	for outpoint, entry := range virtual.utxoSet.utxoCollection {
		diff.toRemove[outpoint] = entry
	}
	genesisPastUTXO := UTXOSet(NewDiffUTXOSet(virtual.utxoSet, diff))
	return genesisPastUTXO
}

func (node *blockNode) fetchBlueBlocks(db database.DB) ([]*util.Block, error) {
	blueBlocks := make([]*util.Block, len(node.blues))
	err := db.View(func(dbTx database.Tx) error {
		for i, blueBlockNode := range node.blues {
			blueBlock, err := dbFetchBlockByNode(dbTx, blueBlockNode)
			if err != nil {
				return err
			}

			blueBlocks[i] = blueBlock
		}

		return nil
	})
	return blueBlocks, err
}

// applyBlueBlocks adds all transactions in the blue blocks to the selectedParent's UTXO set
// Purposefully ignoring failures - these are just unaccepted transactions
// Writing down which transactions were accepted or not in txsAcceptanceData
func (node *blockNode) applyBlueBlocks(acceptedSelectedParentUTXO UTXOSet, selectedParentAcceptanceData []TxAcceptanceData, blueBlocks []*util.Block) (
	pastUTXO UTXOSet, multiBlockTxsAcceptanceData MultiBlockTxsAcceptanceData, err error) {

	pastUTXO = acceptedSelectedParentUTXO
	multiBlockTxsAcceptanceData = MultiBlockTxsAcceptanceData{BlockTxsAcceptanceData{
		BlockHash:        *node.selectedParent.hash,
		TxAcceptanceData: selectedParentAcceptanceData,
	}}

	// Add blueBlocks to multiBlockTxsAcceptanceData in topological order. This
	// is so that anyone who iterates over it would process blocks (and transactions)
	// in their order of appearance in the DAG.
	// We skip the selected parent, because we calculated its UTXO in acceptSelectedParentTransactions.
	for i := 1; i < len(blueBlocks); i++ {
		blueBlock := blueBlocks[i]
		transactions := blueBlock.Transactions()
		blockTxsAcceptanceData := BlockTxsAcceptanceData{
			BlockHash:        *blueBlock.Hash(),
			TxAcceptanceData: make([]TxAcceptanceData, len(transactions)),
		}
		for i, tx := range blueBlock.Transactions() {
			var isAccepted bool
			// Coinbase transaction outputs are added to the UTXO
			// only if they are in the selected parent chain.
			if tx.IsCoinBase() {
				isAccepted = false
			} else {
				isAccepted, err = pastUTXO.AddTx(tx.MsgTx(), node.blueScore)
				if err != nil {
					return nil, nil, err
				}
			}
			blockTxsAcceptanceData.TxAcceptanceData[i] = TxAcceptanceData{Tx: tx, IsAccepted: isAccepted}
		}
		multiBlockTxsAcceptanceData = append(multiBlockTxsAcceptanceData, blockTxsAcceptanceData)
	}

	return pastUTXO, multiBlockTxsAcceptanceData, nil
}

// updateParents adds this block to the children sets of its parents
// and updates the diff of any parent whose DiffChild is this block
func (node *blockNode) updateParents(dag *BlockDAG, newBlockUTXO UTXOSet) error {
	node.updateParentsChildren()
	return node.updateParentsDiffs(dag, newBlockUTXO)
}

// updateParentsDiffs updates the diff of any parent whose DiffChild is this block
func (node *blockNode) updateParentsDiffs(dag *BlockDAG, newBlockUTXO UTXOSet) error {
	virtualDiffFromNewBlock, err := dag.virtual.utxoSet.diffFrom(newBlockUTXO)
	if err != nil {
		return err
	}

	err = dag.utxoDiffStore.setBlockDiff(node, virtualDiffFromNewBlock)
	if err != nil {
		return err
	}

	for parent := range node.parents {
		diffChild, err := dag.utxoDiffStore.diffChildByNode(parent)
		if err != nil {
			return err
		}
		if diffChild == nil {
			parentUTXO, err := dag.restoreUTXO(parent)
			if err != nil {
				return err
			}
			err = dag.utxoDiffStore.setBlockDiffChild(parent, node)
			if err != nil {
				return err
			}
			diff, err := newBlockUTXO.diffFrom(parentUTXO)
			if err != nil {
				return err
			}
			err = dag.utxoDiffStore.setBlockDiff(parent, diff)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// pastUTXO returns the UTXO of a given block's past
// To save traversals over the blue blocks, it also returns the transaction acceptance data for
// all blue blocks
func (dag *BlockDAG) pastUTXO(node *blockNode) (
	pastUTXO UTXOSet, bluesTxsAcceptanceData MultiBlockTxsAcceptanceData, err error) {

	if node.isGenesis() {
		return genesisPastUTXO(dag.virtual), MultiBlockTxsAcceptanceData{}, nil
	}
	selectedParentUTXO, err := dag.restoreUTXO(node.selectedParent)
	if err != nil {
		return nil, nil, err
	}

	blueBlocks, err := node.fetchBlueBlocks(dag.db)
	if err != nil {
		return nil, nil, err
	}

	selectedParent := blueBlocks[0]
	acceptedSelectedParentUTXO, selectedParentAcceptanceData, err := node.acceptSelectedParentTransactions(selectedParent, selectedParentUTXO)
	if err != nil {
		return nil, nil, err
	}

	return node.applyBlueBlocks(acceptedSelectedParentUTXO, selectedParentAcceptanceData, blueBlocks)
}

func (node *blockNode) acceptSelectedParentTransactions(selectedParent *util.Block, selectedParentUTXO UTXOSet) (acceptedSelectedParentUTXO UTXOSet, txAcceptanceData []TxAcceptanceData, err error) {
	diff := NewUTXODiff()
	txAcceptanceData = make([]TxAcceptanceData, len(selectedParent.Transactions()))
	for i, tx := range selectedParent.Transactions() {
		txAcceptanceData[i] = TxAcceptanceData{
			Tx:         tx,
			IsAccepted: true,
		}
		acceptanceDiff, err := selectedParentUTXO.diffFromAcceptedTx(tx.MsgTx(), node.blueScore)
		if err != nil {
			return nil, nil, err
		}
		diff, err = diff.WithDiff(acceptanceDiff)
		if err != nil {
			return nil, nil, err
		}
	}
	acceptedSelectedParentUTXO, err = selectedParentUTXO.WithDiff(diff)
	if err != nil {
		return nil, nil, err
	}
	return acceptedSelectedParentUTXO, txAcceptanceData, nil
}

// restoreUTXO restores the UTXO of a given block from its diff
func (dag *BlockDAG) restoreUTXO(node *blockNode) (UTXOSet, error) {
	stack := []*blockNode{}

	// Iterate over the chain of diff-childs from node till virtual and add them
	// all into a stack
	for current := node; current != nil; {
		stack = append(stack, current)
		var err error
		current, err = dag.utxoDiffStore.diffChildByNode(current)
		if err != nil {
			return nil, err
		}
	}

	// Start with the top item in the stack, going over it top-to-bottom,
	// applying the UTXO-diff one-by-one.
	topNode, stack := stack[len(stack)-1], stack[:len(stack)-1] // pop the top item in the stack
	topNodeDiff, err := dag.utxoDiffStore.diffByNode(topNode)
	if err != nil {
		return nil, err
	}
	accumulatedDiff := topNodeDiff.clone()

	for i := len(stack) - 1; i >= 0; i-- {
		diff, err := dag.utxoDiffStore.diffByNode(stack[i])
		if err != nil {
			return nil, err
		}
		// Use WithDiffInPlace, otherwise copying the diffs again and again create a polynomial overhead
		err = accumulatedDiff.WithDiffInPlace(diff)
		if err != nil {
			return nil, err
		}
	}

	return NewDiffUTXOSet(dag.virtual.utxoSet, accumulatedDiff), nil
}

// updateTipsUTXO builds and applies new diff UTXOs for all the DAG's tips
func updateTipsUTXO(dag *BlockDAG, virtualUTXO UTXOSet) error {
	for tip := range dag.virtual.parents {
		tipUTXO, err := dag.restoreUTXO(tip)
		if err != nil {
			return err
		}
		diff, err := virtualUTXO.diffFrom(tipUTXO)
		if err != nil {
			return err
		}
		err = dag.utxoDiffStore.setBlockDiff(tip, diff)
		if err != nil {
			return err
		}
	}

	return nil
}

// isCurrent returns whether or not the DAG believes it is current. Several
// factors are used to guess, but the key factors that allow the DAG to
// believe it is current are:
//  - Latest block has a timestamp newer than 24 hours ago
//
// This function MUST be called with the DAG state lock held (for reads).
func (dag *BlockDAG) isCurrent() bool {
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
	dagTime := time.Unix(dagTimestamp, 0)
	return dag.Now().Sub(dagTime) <= isDAGCurrentMaxDiff
}

// Now returns the adjusted time according to
// dag.timeSource. See TimeSource.Now for
// more details.
func (dag *BlockDAG) Now() time.Time {
	return dag.timeSource.Now()
}

// IsCurrent returns whether or not the DAG believes it is current. Several
// factors are used to guess, but the key factors that allow the DAG to
// believe it is current are:
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

// SelectedTipHash returns the hash of the current selected tip for the DAG.
// It will return nil if there is no tip.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) SelectedTipHash() *daghash.Hash {
	selectedTip := dag.selectedTip()
	if selectedTip == nil {
		return nil
	}

	return selectedTip.hash
}

// UTXOSet returns the DAG's UTXO set
func (dag *BlockDAG) UTXOSet() *FullUTXOSet {
	return dag.virtual.utxoSet
}

// CalcPastMedianTime returns the past median time of the DAG.
func (dag *BlockDAG) CalcPastMedianTime() time.Time {
	return dag.virtual.tips().bluest().PastMedianTime(dag)
}

// GetUTXOEntry returns the requested unspent transaction output. The returned
// instance must be treated as immutable since it is shared by all callers.
//
// This function is safe for concurrent access. However, the returned entry (if
// any) is NOT.
func (dag *BlockDAG) GetUTXOEntry(outpoint wire.Outpoint) (*UTXOEntry, bool) {
	return dag.virtual.utxoSet.get(outpoint)
}

// BlueScoreByBlockHash returns the blue score of a block with the given hash.
func (dag *BlockDAG) BlueScoreByBlockHash(hash *daghash.Hash) (uint64, error) {
	node := dag.index.LookupNode(hash)
	if node == nil {
		return 0, errors.Errorf("block %s is unknown", hash)
	}

	return node.blueScore, nil
}

// BlockConfirmationsByHash returns the confirmations number for a block with the
// given hash. See blockConfirmations for further details.
//
// This function is safe for concurrent access
func (dag *BlockDAG) BlockConfirmationsByHash(hash *daghash.Hash) (uint64, error) {
	dag.dagLock.RLock()
	defer dag.dagLock.RUnlock()

	return dag.BlockConfirmationsByHashNoLock(hash)
}

// BlockConfirmationsByHashNoLock is lock free version of BlockConfirmationsByHash
//
// This function is unsafe for concurrent access.
func (dag *BlockDAG) BlockConfirmationsByHashNoLock(hash *daghash.Hash) (uint64, error) {
	if hash.IsEqual(&daghash.ZeroHash) {
		return 0, nil
	}

	node := dag.index.LookupNode(hash)
	if node == nil {
		return 0, errors.Errorf("block %s is unknown", hash)
	}

	return dag.blockConfirmations(node)
}

// UTXOConfirmations returns the confirmations for the given outpoint, if it exists
// in the DAG's UTXO set.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) UTXOConfirmations(outpoint *wire.Outpoint) (uint64, bool) {
	dag.dagLock.RLock()
	defer dag.dagLock.RUnlock()

	utxoEntry, ok := dag.GetUTXOEntry(*outpoint)
	if !ok {
		return 0, false
	}
	confirmations := dag.SelectedTipBlueScore() - utxoEntry.BlockBlueScore() + 1

	return confirmations, true
}

// UTXOCommitment returns a commitment to the dag's current UTXOSet
func (dag *BlockDAG) UTXOCommitment() string {
	return dag.UTXOSet().UTXOMultiset.Hash().String()
}

// blockConfirmations returns the current confirmations number of the given node
// The confirmations number is defined as follows:
// * If the node is in the selected tip red set	-> 0
// * If the node is the selected tip			-> 1
// * Otherwise									-> selectedTip.blueScore - acceptingBlock.blueScore + 2
func (dag *BlockDAG) blockConfirmations(node *blockNode) (uint64, error) {
	acceptingBlock, err := dag.acceptingBlock(node)
	if err != nil {
		return 0, err
	}

	// if acceptingBlock is nil, the node is red
	if acceptingBlock == nil {
		return 0, nil
	}

	return dag.selectedTip().blueScore - acceptingBlock.blueScore + 1, nil
}

// acceptingBlock finds the node in the selected-parent chain that had accepted
// the given node
func (dag *BlockDAG) acceptingBlock(node *blockNode) (*blockNode, error) {
	// Return an error if the node is the virtual block
	if node == &dag.virtual.blockNode {
		return nil, errors.New("cannot get acceptingBlock for virtual")
	}

	// If the node is a chain-block itself, the accepting block is its chain-child
	isNodeInSelectedParentChain, err := dag.IsInSelectedParentChain(node.hash)
	if err != nil {
		return nil, err
	}
	if isNodeInSelectedParentChain {
		if len(node.children) == 0 {
			// If the node is the selected tip, it doesn't have an accepting block
			return nil, nil
		}
		for child := range node.children {
			isChildInSelectedParentChain, err := dag.IsInSelectedParentChain(child.hash)
			if err != nil {
				return nil, err
			}
			if isChildInSelectedParentChain {
				return child, nil
			}
		}
		return nil, errors.Errorf("chain block %s does not have a chain child", node.hash)
	}

	// Find the only chain block that may contain the node in its blues
	candidateAcceptingBlock := dag.oldestChainBlockWithBlueScoreGreaterThan(node.blueScore)

	// if no candidate is found, it means that the node has same or more
	// blue score than the selected tip and is found in its anticone, so
	// it doesn't have an accepting block
	if candidateAcceptingBlock == nil {
		return nil, nil
	}

	// candidateAcceptingBlock is the accepting block only if it actually contains
	// the node in its blues
	for _, blue := range candidateAcceptingBlock.blues {
		if blue == node {
			return candidateAcceptingBlock, nil
		}
	}

	// Otherwise, the node is red or in the selected tip anticone, and
	// doesn't have an accepting block
	return nil, nil
}

// oldestChainBlockWithBlueScoreGreaterThan finds the oldest chain block with a blue score
// greater than blueScore. If no such block exists, this method returns nil
func (dag *BlockDAG) oldestChainBlockWithBlueScoreGreaterThan(blueScore uint64) *blockNode {
	chainBlockIndex, ok := util.SearchSlice(len(dag.virtual.selectedParentChainSlice), func(i int) bool {
		selectedPathNode := dag.virtual.selectedParentChainSlice[i]
		return selectedPathNode.blueScore > blueScore
	})
	if !ok {
		return nil
	}
	return dag.virtual.selectedParentChainSlice[chainBlockIndex]
}

// IsInSelectedParentChain returns whether or not a block hash is found in the selected
// parent chain. Note that this method returns an error if the given blockHash does not
// exist within the block index.
//
// This method MUST be called with the DAG lock held
func (dag *BlockDAG) IsInSelectedParentChain(blockHash *daghash.Hash) (bool, error) {
	blockNode := dag.index.LookupNode(blockHash)
	if blockNode == nil {
		str := fmt.Sprintf("block %s is not in the DAG", blockHash)
		return false, errNotInDAG(str)
	}
	return dag.virtual.selectedParentChainSet.contains(blockNode), nil
}

// SelectedParentChain returns the selected parent chain starting from blockHash (exclusive)
// up to the virtual (exclusive). If blockHash is nil then the genesis block is used. If
// blockHash is not within the select parent chain, go down its own selected parent chain,
// while collecting each block hash in removedChainHashes, until reaching a block within
// the main selected parent chain.
//
// This method MUST be called with the DAG lock held
func (dag *BlockDAG) SelectedParentChain(blockHash *daghash.Hash) ([]*daghash.Hash, []*daghash.Hash, error) {
	if blockHash == nil {
		blockHash = dag.genesis.hash
	}
	if !dag.IsInDAG(blockHash) {
		return nil, nil, errors.Errorf("blockHash %s does not exist in the DAG", blockHash)
	}

	// If blockHash is not in the selected parent chain, go down its selected parent chain
	// until we find a block that is in the main selected parent chain.
	var removedChainHashes []*daghash.Hash
	isBlockInSelectedParentChain, err := dag.IsInSelectedParentChain(blockHash)
	if err != nil {
		return nil, nil, err
	}
	for !isBlockInSelectedParentChain {
		removedChainHashes = append(removedChainHashes, blockHash)

		node := dag.index.LookupNode(blockHash)
		blockHash = node.selectedParent.hash

		isBlockInSelectedParentChain, err = dag.IsInSelectedParentChain(blockHash)
		if err != nil {
			return nil, nil, err
		}
	}

	// Find the index of the blockHash in the selectedParentChainSlice
	blockHashIndex := len(dag.virtual.selectedParentChainSlice) - 1
	for blockHashIndex >= 0 {
		node := dag.virtual.selectedParentChainSlice[blockHashIndex]
		if node.hash.IsEqual(blockHash) {
			break
		}
		blockHashIndex--
	}

	// Copy all the addedChainHashes starting from blockHashIndex (exclusive)
	addedChainHashes := make([]*daghash.Hash, len(dag.virtual.selectedParentChainSlice)-blockHashIndex-1)
	for i, node := range dag.virtual.selectedParentChainSlice[blockHashIndex+1:] {
		addedChainHashes[i] = node.hash
	}

	return removedChainHashes, addedChainHashes, nil
}

// SelectedTipBlueScore returns the blue score of the selected tip.
func (dag *BlockDAG) SelectedTipBlueScore() uint64 {
	return dag.selectedTip().blueScore
}

// VirtualBlueScore returns the blue score of the current virtual block
func (dag *BlockDAG) VirtualBlueScore() uint64 {
	return dag.virtual.blueScore
}

// BlockCount returns the number of blocks in the DAG
func (dag *BlockDAG) BlockCount() uint64 {
	return dag.blockCount
}

// TipHashes returns the hashes of the DAG's tips
func (dag *BlockDAG) TipHashes() []*daghash.Hash {
	return dag.virtual.tips().hashes()
}

// CurrentBits returns the bits of the tip with the lowest bits, which also means it has highest difficulty.
func (dag *BlockDAG) CurrentBits() uint32 {
	tips := dag.virtual.tips()
	minBits := uint32(math.MaxUint32)
	for tip := range tips {
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
		err := errors.Errorf("block %s is not known", hash)
		return &wire.BlockHeader{}, err
	}

	return node.Header(), nil
}

// ChildHashesByHash returns the child hashes of the block with the given hash in the
// DAG.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) ChildHashesByHash(hash *daghash.Hash) ([]*daghash.Hash, error) {
	node := dag.index.LookupNode(hash)
	if node == nil {
		str := fmt.Sprintf("block %s is not in the DAG", hash)
		return nil, errNotInDAG(str)

	}

	return node.children.hashes(), nil
}

// SelectedParentHash returns the selected parent hash of the block with the given hash in the
// DAG.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) SelectedParentHash(blockHash *daghash.Hash) (*daghash.Hash, error) {
	node := dag.index.LookupNode(blockHash)
	if node == nil {
		str := fmt.Sprintf("block %s is not in the DAG", blockHash)
		return nil, errNotInDAG(str)

	}

	if node.selectedParent == nil {
		return nil, nil
	}
	return node.selectedParent.hash, nil
}

// antiPastHashesBetween returns the hashes of the blocks between the
// lowHash's antiPast and highHash's antiPast, or up to the provided
// max number of block hashes.
//
// This function MUST be called with the DAG state lock held (for reads).
func (dag *BlockDAG) antiPastHashesBetween(lowHash, highHash *daghash.Hash, maxHashes uint64) ([]*daghash.Hash, error) {
	nodes, err := dag.antiPastBetween(lowHash, highHash, maxHashes)
	if err != nil {
		return nil, err
	}
	hashes := make([]*daghash.Hash, len(nodes))
	for i, node := range nodes {
		hashes[i] = node.hash
	}
	return hashes, nil
}

// antiPastBetween returns the blockNodes between the lowHash's antiPast
// and highHash's antiPast, or up to the provided max number of blocks.
//
// This function MUST be called with the DAG state lock held (for reads).
func (dag *BlockDAG) antiPastBetween(lowHash, highHash *daghash.Hash, maxEntries uint64) ([]*blockNode, error) {
	lowNode := dag.index.LookupNode(lowHash)
	if lowNode == nil {
		return nil, errors.Errorf("Couldn't find low hash %s", lowHash)
	}
	highNode := dag.index.LookupNode(highHash)
	if highNode == nil {
		return nil, errors.Errorf("Couldn't find high hash %s", highHash)
	}
	if lowNode.blueScore >= highNode.blueScore {
		return nil, errors.Errorf("Low hash blueScore >= high hash blueScore (%d >= %d)",
			lowNode.blueScore, highNode.blueScore)
	}

	// In order to get no more then maxEntries blocks from the
	// future of the lowNode (including itself), we iterate the
	// selected parent chain of the highNode and stop once we reach
	// highNode.blueScore-lowNode.blueScore+1 <= maxEntries. That
	// stop point becomes the new highNode.
	// Using blueScore as an approximation is considered to be
	// fairly accurate because we presume that most DAG blocks are
	// blue.
	for highNode.blueScore-lowNode.blueScore+1 > maxEntries {
		highNode = highNode.selectedParent
	}

	// Collect every node in highNode's past (including itself) but
	// NOT in the lowNode's past (excluding itself) into an up-heap
	// (a heap sorted by blueScore from lowest to greatest).
	visited := newBlockSet()
	candidateNodes := newUpHeap()
	queue := newDownHeap()
	queue.Push(highNode)
	for queue.Len() > 0 {
		current := queue.pop()
		if visited.contains(current) {
			continue
		}
		visited.add(current)
		isCurrentAncestorOfLowNode, err := dag.isAncestorOf(current, lowNode)
		if err != nil {
			return nil, err
		}
		if isCurrentAncestorOfLowNode {
			continue
		}
		candidateNodes.Push(current)
		for parent := range current.parents {
			queue.Push(parent)
		}
	}

	// Pop candidateNodes into a slice. Since candidateNodes is
	// an up-heap, it's guaranteed to be ordered from low to high
	nodesLen := int(maxEntries)
	if candidateNodes.Len() < nodesLen {
		nodesLen = candidateNodes.Len()
	}
	nodes := make([]*blockNode, nodesLen)
	for i := 0; i < nodesLen; i++ {
		nodes[i] = candidateNodes.pop()
	}
	return nodes, nil
}

// AntiPastHashesBetween returns the hashes of the blocks between the
// lowHash's antiPast and highHash's antiPast, or up to the provided
// max number of block hashes.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) AntiPastHashesBetween(lowHash, highHash *daghash.Hash, maxHashes uint64) ([]*daghash.Hash, error) {
	dag.dagLock.RLock()
	defer dag.dagLock.RUnlock()
	hashes, err := dag.antiPastHashesBetween(lowHash, highHash, maxHashes)
	if err != nil {
		return nil, err
	}
	return hashes, nil
}

// antiPastHeadersBetween returns the headers of the blocks between the
// lowHash's antiPast and highHash's antiPast, or up to the provided
// max number of block headers.
//
// This function MUST be called with the DAG state lock held (for reads).
func (dag *BlockDAG) antiPastHeadersBetween(lowHash, highHash *daghash.Hash, maxHeaders uint64) ([]*wire.BlockHeader, error) {
	nodes, err := dag.antiPastBetween(lowHash, highHash, maxHeaders)
	if err != nil {
		return nil, err
	}
	headers := make([]*wire.BlockHeader, len(nodes))
	for i, node := range nodes {
		headers[i] = node.Header()
	}
	return headers, nil
}

// GetTopHeaders returns the top wire.MaxBlockHeadersPerMsg block headers ordered by blue score.
func (dag *BlockDAG) GetTopHeaders(highHash *daghash.Hash, maxHeaders uint64) ([]*wire.BlockHeader, error) {
	highNode := &dag.virtual.blockNode
	if highHash != nil {
		highNode = dag.index.LookupNode(highHash)
		if highNode == nil {
			return nil, errors.Errorf("Couldn't find the high hash %s in the dag", highHash)
		}
	}
	headers := make([]*wire.BlockHeader, 0, highNode.blueScore)
	queue := newDownHeap()
	queue.pushSet(highNode.parents)

	visited := newBlockSet()
	for i := uint32(0); queue.Len() > 0 && uint64(len(headers)) < maxHeaders; i++ {
		current := queue.pop()
		if !visited.contains(current) {
			visited.add(current)
			headers = append(headers, current.Header())
			queue.pushSet(current.parents)
		}
	}
	return headers, nil
}

// Lock locks the DAG's UTXO set for writing.
func (dag *BlockDAG) Lock() {
	dag.dagLock.Lock()
}

// Unlock unlocks the DAG's UTXO set for writing.
func (dag *BlockDAG) Unlock() {
	dag.dagLock.Unlock()
}

// RLock locks the DAG's UTXO set for reading.
func (dag *BlockDAG) RLock() {
	dag.dagLock.RLock()
}

// RUnlock unlocks the DAG's UTXO set for reading.
func (dag *BlockDAG) RUnlock() {
	dag.dagLock.RUnlock()
}

// AntiPastHeadersBetween returns the headers of the blocks between the
// lowHash's antiPast and highHash's antiPast, or up to
// wire.MaxBlockHeadersPerMsg block headers.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) AntiPastHeadersBetween(lowHash, highHash *daghash.Hash, maxHeaders uint64) ([]*wire.BlockHeader, error) {
	dag.dagLock.RLock()
	defer dag.dagLock.RUnlock()
	headers, err := dag.antiPastHeadersBetween(lowHash, highHash, maxHeaders)
	if err != nil {
		return nil, err
	}
	return headers, nil
}

// SubnetworkID returns the node's subnetwork ID
func (dag *BlockDAG) SubnetworkID() *subnetworkid.SubnetworkID {
	return dag.subnetworkID
}

func (dag *BlockDAG) addDelayedBlock(block *util.Block, delay time.Duration) error {
	processTime := dag.Now().Add(delay)
	log.Debugf("Adding block to delayed blocks queue (block hash: %s, process time: %s)", block.Hash().String(), processTime)
	delayedBlock := &delayedBlock{
		block:       block,
		processTime: processTime,
	}

	dag.delayedBlocks[*block.Hash()] = delayedBlock
	dag.delayedBlocksQueue.Push(delayedBlock)
	return dag.processDelayedBlocks()
}

// processDelayedBlocks loops over all delayed blocks and processes blocks which are due.
// This method is invoked after processing a block (ProcessBlock method).
func (dag *BlockDAG) processDelayedBlocks() error {
	// Check if the delayed block with the earliest process time should be processed
	for dag.delayedBlocksQueue.Len() > 0 {
		earliestDelayedBlockProcessTime := dag.peekDelayedBlock().processTime
		if earliestDelayedBlockProcessTime.After(dag.Now()) {
			break
		}
		delayedBlock := dag.popDelayedBlock()
		_, _, err := dag.processBlockNoLock(delayedBlock.block, BFAfterDelay)
		if err != nil {
			log.Errorf("Error while processing delayed block (block %s)", delayedBlock.block.Hash().String())
			// Rule errors should not be propagated as they refer only to the delayed block,
			// while this function runs in the context of another block
			if !errors.As(err, &RuleError{}) {
				return err
			}
		}
		log.Debugf("Processed delayed block (block %s)", delayedBlock.block.Hash().String())
	}

	return nil
}

// popDelayedBlock removes the topmost (delayed block with earliest process time) of the queue and returns it.
func (dag *BlockDAG) popDelayedBlock() *delayedBlock {
	delayedBlock := dag.delayedBlocksQueue.pop()
	delete(dag.delayedBlocks, *delayedBlock.block.Hash())
	return delayedBlock
}

func (dag *BlockDAG) peekDelayedBlock() *delayedBlock {
	return dag.delayedBlocksQueue.peek()
}

// IndexManager provides a generic interface that is called when blocks are
// connected to the DAG for the purpose of supporting optional indexes.
type IndexManager interface {
	// Init is invoked during DAG initialize in order to allow the index
	// manager to initialize itself and any indexes it is managing. The
	// channel parameter specifies a channel the caller can close to signal
	// that the process should be interrupted. It can be nil if that
	// behavior is not desired.
	Init(database.DB, *BlockDAG, <-chan struct{}) error

	// ConnectBlock is invoked when a new block has been connected to the
	// DAG.
	ConnectBlock(dbTx database.Tx, block *util.Block, blockID uint64, dag *BlockDAG, acceptedTxsData MultiBlockTxsAcceptanceData, virtualTxsAcceptanceData MultiBlockTxsAcceptanceData) error
}

// Config is a descriptor which specifies the blockDAG instance configuration.
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

	// TimeSource defines the time source to use for things such as
	// block processing and determining whether or not the DAG is current.
	TimeSource TimeSource

	// SigCache defines a signature cache to use when when validating
	// signatures. This is typically most useful when individual
	// transactions are already being validated prior to their inclusion in
	// a block such as what is usually done via a transaction memory pool.
	//
	// This field can be nil if the caller is not interested in using a
	// signature cache.
	SigCache *txscript.SigCache

	// IndexManager defines an index manager to use when initializing the
	// DAG and connecting blocks.
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

	params := config.DAGParams
	targetTimePerBlock := int64(params.TargetTimePerBlock / time.Second)

	index := newBlockIndex(config.DB, params)
	dag := &BlockDAG{
		db:                             config.DB,
		dagParams:                      params,
		timeSource:                     config.TimeSource,
		sigCache:                       config.SigCache,
		indexManager:                   config.IndexManager,
		targetTimePerBlock:             targetTimePerBlock,
		difficultyAdjustmentWindowSize: params.DifficultyAdjustmentWindowSize,
		TimestampDeviationTolerance:    params.TimestampDeviationTolerance,
		powMaxBits:                     util.BigToCompact(params.PowMax),
		index:                          index,
		orphans:                        make(map[daghash.Hash]*orphanBlock),
		prevOrphans:                    make(map[daghash.Hash][]*orphanBlock),
		delayedBlocks:                  make(map[daghash.Hash]*delayedBlock),
		delayedBlocksQueue:             newDelayedBlocksHeap(),
		warningCaches:                  newThresholdCaches(vbNumBits),
		deploymentCaches:               newThresholdCaches(dagconfig.DefinedDeployments),
		blockCount:                     0,
		SubnetworkStore:                newSubnetworkStore(config.DB),
		subnetworkID:                   config.SubnetworkID,
	}

	dag.virtual = newVirtualBlock(dag, nil)
	dag.utxoDiffStore = newUTXODiffStore(dag)
	dag.reachabilityStore = newReachabilityStore(dag)

	// Initialize the DAG state from the passed database. When the db
	// does not yet contain any DAG state, both it and the DAG state
	// will be initialized to contain only the genesis block.
	err := dag.initDAGState()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			err := dag.removeDAGState()
			if err != nil {
				panic(fmt.Sprintf("Couldn't remove the DAG State: %s", err))
			}
		}
	}()

	// Initialize and catch up all of the currently active optional indexes
	// as needed.
	if config.IndexManager != nil {
		err = config.IndexManager.Init(dag.db, dag, config.Interrupt)
		if err != nil {
			return nil, err
		}
	}

	genesis := index.LookupNode(params.GenesisHash)

	if genesis == nil {
		genesisBlock := util.NewBlock(dag.dagParams.GenesisBlock)
		// To prevent the creation of a new err variable unintentionally so the
		// defered function above could read err - declare isOrphan and isDelayed explicitly.
		var isOrphan, isDelayed bool
		isOrphan, isDelayed, err = dag.ProcessBlock(genesisBlock, BFNone)
		if err != nil {
			return nil, err
		}
		if isDelayed {
			return nil, errors.New("Genesis block shouldn't be in the future")
		}
		if isOrphan {
			return nil, errors.New("Genesis block is unexpectedly orphan")
		}
		genesis = index.LookupNode(params.GenesisHash)
	}

	// Save a reference to the genesis block.
	dag.genesis = genesis

	// Initialize rule change threshold state caches.
	err = dag.initThresholdCaches()
	if err != nil {
		return nil, err
	}

	selectedTip := dag.selectedTip()
	log.Infof("DAG state (blue score %d, hash %s)",
		selectedTip.blueScore, selectedTip.hash)

	return dag, nil
}

func (dag *BlockDAG) isKnownDelayedBlock(hash *daghash.Hash) bool {
	_, exists := dag.delayedBlocks[*hash]
	return exists
}
