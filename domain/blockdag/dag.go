// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"fmt"
	"sync"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/util/subnetworkid"

	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/txscript"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
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

	utxoDiffStore *utxoDiffStore
	multisetStore *multisetStore

	reachabilityTree *reachabilityTree

	recentBlockProcessingTimestamps []mstime.Time
	startTime                       mstime.Time

	tips blockSet

	// validTips is a set of blocks with the status "valid", which have no valid descendants.
	// Note that some validTips might not be actual tips.
	validTips blockSet
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

// Lock locks the DAG for writing.
func (dag *BlockDAG) Lock() {
	dag.dagLock.Lock()
}

// Unlock unlocks the DAG for writing.
func (dag *BlockDAG) Unlock() {
	dag.dagLock.Unlock()
}

// RLock locks the DAG for reading.
func (dag *BlockDAG) RLock() {
	dag.dagLock.RLock()
}

// RUnlock unlocks the DAG for reading.
func (dag *BlockDAG) RUnlock() {
	dag.dagLock.RUnlock()
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
func (dag *BlockDAG) SelectedTipHeader() *appmessage.BlockHeader {
	dag.dagLock.RLock()
	defer dag.dagLock.RUnlock()
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
	return dag.virtual.selectedParent.PastMedianTime()
}

// GetUTXOEntry returns the requested unspent transaction output. The returned
// instance must be treated as immutable since it is shared by all callers.
//
// This function is safe for concurrent access. However, the returned entry (if
// any) is NOT.
func (dag *BlockDAG) GetUTXOEntry(outpoint appmessage.Outpoint) (*UTXOEntry, bool) {
	dag.RLock()
	defer dag.RUnlock()
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

// SelectedTipBlueScore returns the blue score of the selected tip.
func (dag *BlockDAG) SelectedTipBlueScore() uint64 {
	return dag.selectedTip().blueScore
}

// VirtualBlueHashes returns the blue of the current virtual block
func (dag *BlockDAG) VirtualBlueHashes() []*daghash.Hash {
	dag.RLock()
	defer dag.RUnlock()
	hashes := make([]*daghash.Hash, len(dag.virtual.blues))
	for i, blue := range dag.virtual.blues {
		hashes[i] = blue.hash
	}
	return hashes
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
	return dag.tips.hashes()
}

// ValidTipHashes returns the hashes of the DAG's valid tips
func (dag *BlockDAG) ValidTipHashes() []*daghash.Hash {
	return dag.validTips.hashes()
}

// VirtualParentHashes returns the hashes of the virtual block's parents
func (dag *BlockDAG) VirtualParentHashes() []*daghash.Hash {
	return dag.virtual.parents.hashes()
}

// HeaderByHash returns the block header identified by the given hash or an
// error if it doesn't exist.
func (dag *BlockDAG) HeaderByHash(hash *daghash.Hash) (*appmessage.BlockHeader, error) {
	node, ok := dag.index.LookupNode(hash)
	if !ok {
		err := errors.Errorf("block %s is not known", hash)
		return &appmessage.BlockHeader{}, err
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

// isInPast returns true if `node` is in the past of `other`
//
// Note: this method will return true if `node == other`
func (dag *BlockDAG) isInPast(node *blockNode, other *blockNode) (bool, error) {
	return dag.reachabilityTree.isInPast(node, other)
}

// isInPastOfAny returns true if `node` is in the past of any of `others`
//
// Note: this method will return true if `node` is in `others`
func (dag *BlockDAG) isInPastOfAny(node *blockNode, others blockSet) (bool, error) {
	for other := range others {
		isInPast, err := dag.isInPast(node, other)
		if err != nil {
			return false, err
		}
		if isInPast {
			return true, nil
		}
	}

	return false, nil
}

// isInPastOfAny returns true if any one of `nodes` is in the past of any of `others`
//
// Note: this method will return true if `other` is in `nodes`
func (dag *BlockDAG) isAnyInPastOf(nodes blockSet, other *blockNode) (bool, error) {
	for node := range nodes {
		isInPast, err := dag.isInPast(node, other)
		if err != nil {
			return false, err
		}
		if isInPast {
			return true, nil
		}
	}

	return false, nil
}

// GetTopHeaders returns the top appmessage.MaxBlockHeadersPerMsg block headers ordered by blue score.
func (dag *BlockDAG) GetTopHeaders(highHash *daghash.Hash, maxHeaders uint64) ([]*appmessage.BlockHeader, error) {
	highNode := dag.virtual.blockNode
	if highHash != nil {
		var ok bool
		highNode, ok = dag.index.LookupNode(highHash)
		if !ok {
			return nil, errors.Errorf("Couldn't find the high hash %s in the dag", highHash)
		}
	}
	headers := make([]*appmessage.BlockHeader, 0, highNode.blueScore)
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
	return dag.index.BlockNodeStatus(node).KnownInvalid()
}

func (dag *BlockDAG) addTip(tip *blockNode) (
	didVirtualParentsChange bool, chainUpdates *selectedParentChainUpdates, err error) {

	newTips := dag.tips.clone()
	for parent := range tip.parents {
		newTips.remove(parent)
	}

	newTips.add(tip)

	return dag.setTips(newTips)
}

func (dag *BlockDAG) setTips(newTips blockSet) (
	didVirtualParentsChange bool, chainUpdates *selectedParentChainUpdates, err error) {

	didVirtualParentsChange, chainUpdates, err = dag.updateVirtualParents(newTips, dag.virtual.finalityPoint())
	if err != nil {
		return false, nil, err
	}

	dag.tips = newTips

	return didVirtualParentsChange, chainUpdates, nil
}

func (dag *BlockDAG) updateVirtualParents(newTips blockSet, finalityPoint *blockNode) (
	didVirtualParentsChange bool, chainUpdates *selectedParentChainUpdates, err error) {

	var newVirtualParents blockSet
	// If only genesis is the newTips - we are still initializing the DAG and not all structures required
	// for calling dag.selectVirtualParents have been initialized yet.
	// Specifically - this function would be called with finalityPoint = dag.virtual.finalityPoint(), which has
	// not been initialized to anything real yet.
	//
	// Therefore, in this case - simply pick genesis as virtual's only parent
	if newTips.isOnlyGenesis() {
		newVirtualParents = newTips
	} else {
		newVirtualParents, err = dag.selectVirtualParents(newTips, finalityPoint)
		if err != nil {
			return false, nil, err
		}
	}

	oldVirtualParents := dag.virtual.parents
	didVirtualParentsChange = !oldVirtualParents.isEqual(newVirtualParents)

	if !didVirtualParentsChange {
		return false, &selectedParentChainUpdates{}, nil
	}

	oldSelectedParent := dag.virtual.selectedParent
	dag.virtual.blockNode, _ = dag.newBlockNode(nil, newVirtualParents)
	chainUpdates = dag.virtual.updateSelectedParentSet(oldSelectedParent)

	return didVirtualParentsChange, chainUpdates, nil
}

func (dag *BlockDAG) addValidTip(newValidTip *blockNode) error {
	newValidTips := dag.validTips.clone()
	for validTip := range dag.validTips {
		// We use isInPastOfAny on newValidTip.parents instead of
		// isInPast on newValidTip because newValidTip does not
		// yet have reachability data associated with it.
		isInPastOfNewValidTip, err := dag.isInPastOfAny(validTip, newValidTip.parents)
		if err != nil {
			return err
		}
		if isInPastOfNewValidTip {
			newValidTips.remove(validTip)
		}
	}

	newValidTips.add(newValidTip)

	dag.validTips = newValidTips
	return nil
}

func (dag *BlockDAG) selectVirtualParents(tips blockSet, finalityPoint *blockNode) (blockSet, error) {
	selected := newBlockSet()

	candidatesHeap := newDownHeap()
	candidatesHeap.pushSet(tips)

	// If the first candidate has been disqualified from the chain or violates finality -
	// it cannot be virtual's parent, since it will make it virtual's selectedParent - disqualifying virtual itself.
	// Therefore, in such a case we remove it from the list of virtual parent candidates, and replace with
	// its parents that have no disqualified children
	disqualifiedCandidates := newBlockSet()
	for {
		if candidatesHeap.Len() == 0 {
			return nil, errors.New("virtual has no valid parent candidates")
		}
		selectedParentCandidate := candidatesHeap.pop()

		if dag.index.BlockNodeStatus(selectedParentCandidate) == statusValid {
			isFinalityPointInSelectedParentChain, err := dag.isInSelectedParentChainOf(finalityPoint, selectedParentCandidate)
			if err != nil {
				return nil, err
			}

			if isFinalityPointInSelectedParentChain {
				selected.add(selectedParentCandidate)
				break
			}
		}

		disqualifiedCandidates.add(selectedParentCandidate)

		for parent := range selectedParentCandidate.parents {
			if parent.children.areAllIn(disqualifiedCandidates) {
				candidatesHeap.Push(parent)
			}
		}
	}

	mergeSetSize := 1 // starts counting from 1 because selectedParent is already in the mergeSet

	for len(selected) < appmessage.MaxBlockParents && candidatesHeap.Len() > 0 {
		candidate := candidatesHeap.pop()

		// check that the candidate doesn't increase the virtual's merge set over `mergeSetSizeLimit`
		mergeSetIncrease, err := dag.mergeSetIncrease(candidate, selected)
		if err != nil {
			return nil, err
		}

		if mergeSetSize+mergeSetIncrease > mergeSetSizeLimit {
			continue
		}

		selected.add(candidate)
		mergeSetSize += mergeSetIncrease
	}

	tempVirtual, _ := dag.newBlockNode(nil, selected)

	boundedMergeBreakingParents, err := dag.boundedMergeBreakingParents(tempVirtual)
	if err != nil {
		return nil, err
	}
	selected = selected.subtract(boundedMergeBreakingParents)

	return selected, nil
}

func (dag *BlockDAG) mergeSetIncrease(candidate *blockNode, selected blockSet) (int, error) {
	visited := newBlockSet()
	queue := newDownHeap()
	queue.Push(candidate)
	mergeSetIncrease := 1 // starts with 1 for the candidate itself

	for queue.Len() > 0 {
		current := queue.pop()
		isInPastOfSelected, err := dag.isInPastOfAny(current, selected)
		if err != nil {
			return 0, err
		}
		if isInPastOfSelected {
			continue
		}
		mergeSetIncrease++

		for parent := range current.parents {
			if !visited.contains(parent) {
				visited.add(parent)
				queue.Push(parent)
			}
		}
	}

	return mergeSetIncrease, nil
}

func (dag *BlockDAG) saveState(dbTx *dbaccess.TxContext) error {
	state := &dagState{
		TipHashes:            dag.TipHashes(),
		ValidTipHashes:       dag.ValidTipHashes(),
		VirtualParentsHashes: dag.VirtualParentHashes(),
		LocalSubnetworkID:    dag.subnetworkID,
	}
	return saveDAGState(dbTx, state)
}
