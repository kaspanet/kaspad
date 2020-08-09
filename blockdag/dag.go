// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"fmt"
	"github.com/kaspanet/kaspad/util/mstime"
	"math"
	"sync"

	"github.com/kaspanet/kaspad/dbaccess"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/util/subnetworkid"

	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/txscript"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

// BlockDAG provides functions for working with the kaspa block DAG.
// It includes functionality such as rejecting duplicate blocks, ensuring blocks
// follow all rules, and orphan handling.
type BlockDAG struct {
	// The following fields are set when the instance is created and can't
	// be changed afterwards, so there is no need to protect them with a
	// separate mutex.
	Params          *dagconfig.Params
	databaseContext *dbaccess.DatabaseContext
	timeSource      TimeSource
	sigCache        *txscript.SigCache
	indexManager    IndexManager
	genesis         *blockNode

	// The following fields are calculated based upon the provided DAG
	// parameters. They are also set when the instance is created and
	// can't be changed afterwards, so there is no need to protect them with
	// a separate mutex.
	difficultyAdjustmentWindowSize uint64
	TimestampDeviationTolerance    uint64

	// powMaxBits defines the highest allowed proof of work value for a
	// block in compact form.
	powMaxBits uint32

	// dagLock protects concurrent access to the vast majority of the
	// fields in this struct below this point.
	dagLock sync.RWMutex

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

	// The notifications field stores a slice of callbacks to be executed on
	// certain blockDAG events.
	notificationsLock sync.RWMutex
	notifications     []NotificationCallback

	lastFinalityPoint *blockNode

	utxoDiffStore *utxoDiffStore
	multisetStore *multisetStore

	reachabilityTree *reachabilityTree

	recentBlockProcessingTimestamps []mstime.Time
	startTime                       mstime.Time
}

// New returns a BlockDAG instance using the provided configuration details.
func New(config *Config) (*BlockDAG, error) {
	params := config.DAGParams

	dag := &BlockDAG{
		Params:                         params,
		databaseContext:                config.DatabaseContext,
		timeSource:                     config.TimeSource,
		sigCache:                       config.SigCache,
		indexManager:                   config.IndexManager,
		difficultyAdjustmentWindowSize: params.DifficultyAdjustmentWindowSize,
		TimestampDeviationTolerance:    params.TimestampDeviationTolerance,
		powMaxBits:                     util.BigToCompact(params.PowMax),
		index:                          newBlockIndex(params),
		orphans:                        make(map[daghash.Hash]*orphanBlock),
		prevOrphans:                    make(map[daghash.Hash][]*orphanBlock),
		delayedBlocks:                  make(map[daghash.Hash]*delayedBlock),
		delayedBlocksQueue:             newDelayedBlocksHeap(),
		blockCount:                     0,
		subnetworkID:                   config.SubnetworkID,
		startTime:                      mstime.Now(),
	}

	dag.virtual = newVirtualBlock(dag, nil)
	dag.utxoDiffStore = newUTXODiffStore(dag)
	dag.multisetStore = newMultisetStore(dag)
	dag.reachabilityTree = newReachabilityTree(dag)

	// Initialize the DAG state from the passed database. When the db
	// does not yet contain any DAG state, both it and the DAG state
	// will be initialized to contain only the genesis block.
	err := dag.initDAGState()
	if err != nil {
		return nil, err
	}

	// Initialize and catch up all of the currently active optional indexes
	// as needed.
	if config.IndexManager != nil {
		err = config.IndexManager.Init(dag, dag.databaseContext)
		if err != nil {
			return nil, err
		}
	}

	genesis, ok := dag.index.LookupNode(params.GenesisHash)

	if !ok {
		genesisBlock := util.NewBlock(dag.Params.GenesisBlock)
		isOrphan, isDelayed, err := dag.ProcessBlock(genesisBlock, BFNone)
		if err != nil {
			return nil, err
		}
		if isDelayed {
			return nil, errors.New("genesis block shouldn't be in the future")
		}
		if isOrphan {
			return nil, errors.New("genesis block is unexpectedly orphan")
		}
		genesis, ok = dag.index.LookupNode(params.GenesisHash)
		if !ok {
			return nil, errors.New("genesis is not found in the DAG after it was proccessed")
		}
	}

	// Save a reference to the genesis block.
	dag.genesis = genesis

	selectedTip := dag.selectedTip()
	log.Infof("DAG state (blue score %d, hash %s)",
		selectedTip.blueScore, selectedTip.hash)

	return dag, nil
}

// TxsAcceptedByVirtual retrieves transactions accepted by the current virtual block
//
// This function MUST be called with the DAG read-lock held
func (dag *BlockDAG) TxsAcceptedByVirtual() (MultiBlockTxsAcceptanceData, error) {
	_, _, txsAcceptanceData, err := dag.pastUTXO(&dag.virtual.blockNode)
	return txsAcceptanceData, err
}

// TxsAcceptedByBlockHash retrieves transactions accepted by the given block
//
// This function MUST be called with the DAG read-lock held
func (dag *BlockDAG) TxsAcceptedByBlockHash(blockHash *daghash.Hash) (MultiBlockTxsAcceptanceData, error) {
	node, ok := dag.index.LookupNode(blockHash)
	if !ok {
		return nil, errors.Errorf("Couldn't find block %s", blockHash)
	}
	_, _, txsAcceptanceData, err := dag.pastUTXO(node)
	return txsAcceptanceData, err
}

func (dag *BlockDAG) meldVirtualUTXO(newVirtualUTXODiffSet *DiffUTXOSet) error {
	return newVirtualUTXODiffSet.meldToBase()
}

// checkDoubleSpendsWithBlockPast checks that each block transaction
// has a corresponding UTXO in the block pastUTXO.
func checkDoubleSpendsWithBlockPast(pastUTXO UTXOSet, blockTransactions []*util.Tx) error {
	for _, tx := range blockTransactions {
		if tx.IsCoinBase() {
			continue
		}

		for _, txIn := range tx.MsgTx().TxIn {
			if _, ok := pastUTXO.Get(txIn.PreviousOutpoint); !ok {
				return ruleError(ErrMissingTxOut, fmt.Sprintf("missing transaction "+
					"output %s in the utxo set", txIn.PreviousOutpoint))
			}
		}
	}

	return nil
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

// applyBlueBlocks adds all transactions in the blue blocks to the selectedParent's past UTXO set
// Purposefully ignoring failures - these are just unaccepted transactions
// Writing down which transactions were accepted or not in txsAcceptanceData
func (node *blockNode) applyBlueBlocks(selectedParentPastUTXO UTXOSet, blueBlocks []*util.Block) (
	pastUTXO UTXOSet, multiBlockTxsAcceptanceData MultiBlockTxsAcceptanceData, err error) {

	pastUTXO = selectedParentPastUTXO.(*DiffUTXOSet).cloneWithoutBase()
	multiBlockTxsAcceptanceData = make(MultiBlockTxsAcceptanceData, len(blueBlocks))

	// Add blueBlocks to multiBlockTxsAcceptanceData in topological order. This
	// is so that anyone who iterates over it would process blocks (and transactions)
	// in their order of appearance in the DAG.
	for i := 0; i < len(blueBlocks); i++ {
		blueBlock := blueBlocks[i]
		transactions := blueBlock.Transactions()
		blockTxsAcceptanceData := BlockTxsAcceptanceData{
			BlockHash:        *blueBlock.Hash(),
			TxAcceptanceData: make([]TxAcceptanceData, len(transactions)),
		}
		isSelectedParent := i == 0

		for j, tx := range blueBlock.Transactions() {
			var isAccepted bool

			// Coinbase transaction outputs are added to the UTXO
			// only if they are in the selected parent chain.
			if !isSelectedParent && tx.IsCoinBase() {
				isAccepted = false
			} else {
				isAccepted, err = pastUTXO.AddTx(tx.MsgTx(), node.blueScore)
				if err != nil {
					return nil, nil, err
				}
			}
			blockTxsAcceptanceData.TxAcceptanceData[j] = TxAcceptanceData{Tx: tx, IsAccepted: isAccepted}
		}
		multiBlockTxsAcceptanceData[i] = blockTxsAcceptanceData
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
			parentPastUTXO, err := dag.restorePastUTXO(parent)
			if err != nil {
				return err
			}
			err = dag.utxoDiffStore.setBlockDiffChild(parent, node)
			if err != nil {
				return err
			}
			diff, err := newBlockUTXO.diffFrom(parentPastUTXO)
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
	pastUTXO, selectedParentPastUTXO UTXOSet, bluesTxsAcceptanceData MultiBlockTxsAcceptanceData, err error) {

	if node.isGenesis() {
		return genesisPastUTXO(dag.virtual), nil, MultiBlockTxsAcceptanceData{}, nil
	}

	selectedParentPastUTXO, err = dag.restorePastUTXO(node.selectedParent)
	if err != nil {
		return nil, nil, nil, err
	}

	blueBlocks, err := dag.fetchBlueBlocks(node)
	if err != nil {
		return nil, nil, nil, err
	}

	pastUTXO, bluesTxsAcceptanceData, err = node.applyBlueBlocks(selectedParentPastUTXO, blueBlocks)
	if err != nil {
		return nil, nil, nil, err
	}

	return pastUTXO, selectedParentPastUTXO, bluesTxsAcceptanceData, nil
}

// restorePastUTXO restores the UTXO of a given block from its diff
func (dag *BlockDAG) restorePastUTXO(node *blockNode) (UTXOSet, error) {
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
		// Use withDiffInPlace, otherwise copying the diffs again and again create a polynomial overhead
		err = accumulatedDiff.withDiffInPlace(diff)
		if err != nil {
			return nil, err
		}
	}

	return NewDiffUTXOSet(dag.virtual.utxoSet, accumulatedDiff), nil
}

// updateTipsUTXO builds and applies new diff UTXOs for all the DAG's tips
func updateTipsUTXO(dag *BlockDAG, virtualUTXO UTXOSet) error {
	for tip := range dag.virtual.parents {
		tipPastUTXO, err := dag.restorePastUTXO(tip)
		if err != nil {
			return err
		}
		diff, err := virtualUTXO.diffFrom(tipPastUTXO)
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

// Now returns the adjusted time according to
// dag.timeSource. See TimeSource.Now for
// more details.
func (dag *BlockDAG) Now() mstime.Time {
	return dag.timeSource.Now()
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
func (dag *BlockDAG) CalcPastMedianTime() mstime.Time {
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
	node, ok := dag.index.LookupNode(hash)
	if !ok {
		return 0, errors.Errorf("block %s is unknown", hash)
	}

	return node.blueScore, nil
}

// BluesByBlockHash returns the blues of the block for the given hash.
func (dag *BlockDAG) BluesByBlockHash(hash *daghash.Hash) ([]*daghash.Hash, error) {
	node, ok := dag.index.LookupNode(hash)
	if !ok {
		return nil, errors.Errorf("block %s is unknown", hash)
	}

	hashes := make([]*daghash.Hash, len(node.blues))
	for i, blue := range node.blues {
		hashes[i] = blue.hash
	}

	return hashes, nil
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

	node, ok := dag.index.LookupNode(hash)
	if !ok {
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
	blockNode, ok := dag.index.LookupNode(blockHash)
	if !ok {
		str := fmt.Sprintf("block %s is not in the DAG", blockHash)
		return false, ErrNotInDAG(str)
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

		node, ok := dag.index.LookupNode(blockHash)
		if !ok {
			return nil, nil, errors.Errorf("block %s does not exist in the DAG", blockHash)
		}
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
	node, ok := dag.index.LookupNode(hash)
	if !ok {
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
	node, ok := dag.index.LookupNode(hash)
	if !ok {
		str := fmt.Sprintf("block %s is not in the DAG", hash)
		return nil, ErrNotInDAG(str)

	}

	return node.children.hashes(), nil
}

// SelectedParentHash returns the selected parent hash of the block with the given hash in the
// DAG.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) SelectedParentHash(blockHash *daghash.Hash) (*daghash.Hash, error) {
	node, ok := dag.index.LookupNode(blockHash)
	if !ok {
		str := fmt.Sprintf("block %s is not in the DAG", blockHash)
		return nil, ErrNotInDAG(str)

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
	lowNode, ok := dag.index.LookupNode(lowHash)
	if !ok {
		return nil, errors.Errorf("Couldn't find low hash %s", lowHash)
	}
	highNode, ok := dag.index.LookupNode(highHash)
	if !ok {
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
		isCurrentAncestorOfLowNode, err := dag.isInPast(current, lowNode)
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

func (dag *BlockDAG) isInPast(this *blockNode, other *blockNode) (bool, error) {
	return dag.reachabilityTree.isInPast(this, other)
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
		var ok bool
		highNode, ok = dag.index.LookupNode(highHash)
		if !ok {
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

// ForEachHash runs the given fn on every hash that's currently known to
// the DAG.
//
// This function is NOT safe for concurrent access. It is meant to be
// used either on initialization or when the dag lock is held for reads.
func (dag *BlockDAG) ForEachHash(fn func(hash daghash.Hash) error) error {
	for hash := range dag.index.index {
		err := fn(hash)
		if err != nil {
			return err
		}
	}
	return nil
}

// IsInDAG determines whether a block with the given hash exists in
// the DAG.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) IsInDAG(hash *daghash.Hash) bool {
	return dag.index.HaveBlock(hash)
}

// IsKnownBlock returns whether or not the DAG instance has the block represented
// by the passed hash. This includes checking the various places a block can
// be in, like part of the DAG or the orphan pool.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) IsKnownBlock(hash *daghash.Hash) bool {
	return dag.IsInDAG(hash) || dag.IsKnownOrphan(hash) || dag.isKnownDelayedBlock(hash) || dag.IsKnownInvalid(hash)
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

// IsKnownInvalid returns whether the passed hash is known to be an invalid block.
// Note that if the block is not found this method will return false.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) IsKnownInvalid(hash *daghash.Hash) bool {
	node, ok := dag.index.LookupNode(hash)
	if !ok {
		return false
	}
	return dag.index.NodeStatus(node).KnownInvalid()
}

// isInSelectedParentChainOf returns whether `node` is in the selected parent chain of `other`.
func (dag *BlockDAG) isInSelectedParentChainOf(node *blockNode, other *blockNode) (bool, error) {
	// By definition, a node is not in the selected parent chain of itself.
	if node == other {
		return false, nil
	}

	return dag.reachabilityTree.isReachabilityTreeAncestorOf(node, other)
}
