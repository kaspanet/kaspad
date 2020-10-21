// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"fmt"
	"github.com/kaspanet/kaspad/domain/multiset"
	"sync"

	"github.com/kaspanet/kaspad/domain/blocknode"
	"github.com/kaspanet/kaspad/domain/utxo"
	utxodiffstore "github.com/kaspanet/kaspad/domain/utxo"

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
	DatabaseContext *dbaccess.DatabaseContext
	timeSource      TimeSource
	sigCache        *txscript.SigCache
	indexManager    IndexManager
	genesis         *blocknode.Node

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

	// Index and virtual are related to the memory block Index. They both
	// have their own locks, however they are often also protected by the
	// DAG lock to help prevent logic races when blocks are being processed.

	// Index houses the entire block Index in memory. The block Index is
	// a tree-shaped structure.
	Index *blocknode.Index

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

	UTXODiffStore *utxodiffstore.DiffStore
	multisetStore *multiset.Store

	reachabilityTree *reachabilityTree

	recentBlockProcessingTimestamps []mstime.Time
	startTime                       mstime.Time

	maxUTXOCacheSize uint64
	tips             blocknode.Set

	// validTips is a set of blocks with the status "valid", which have no valid descendants.
	// Note that some validTips might not be actual tips.
	validTips blocknode.Set
}

// New returns a BlockDAG instance using the provided configuration details.
func New(config *Config) (*BlockDAG, error) {
	params := config.DAGParams

	dag := &BlockDAG{
		Params:                         params,
		DatabaseContext:                config.DatabaseContext,
		timeSource:                     config.TimeSource,
		sigCache:                       config.SigCache,
		indexManager:                   config.IndexManager,
		difficultyAdjustmentWindowSize: params.DifficultyAdjustmentWindowSize,
		TimestampDeviationTolerance:    params.TimestampDeviationTolerance,
		powMaxBits:                     util.BigToCompact(params.PowMax),
		Index:                          blocknode.NewIndex(),
		orphans:                        make(map[daghash.Hash]*orphanBlock),
		prevOrphans:                    make(map[daghash.Hash][]*orphanBlock),
		delayedBlocks:                  make(map[daghash.Hash]*delayedBlock),
		delayedBlocksQueue:             newDelayedBlocksHeap(),
		blockCount:                     0,
		subnetworkID:                   config.SubnetworkID,
		startTime:                      mstime.Now(),
		maxUTXOCacheSize:               config.MaxUTXOCacheSize,
	}

	dag.virtual = newVirtualBlock(utxo.NewFullUTXOSetFromContext(dag.DatabaseContext, dag.maxUTXOCacheSize), nil, dag.Now().UnixMilliseconds())
	dag.UTXODiffStore = utxo.NewDiffStore(dag.DatabaseContext)
	dag.multisetStore = multiset.NewStore()
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
		err = config.IndexManager.Init(dag, dag.DatabaseContext)
		if err != nil {
			return nil, err
		}
	}

	genesis, ok := dag.Index.LookupNode(params.GenesisHash)

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
		genesis, ok = dag.Index.LookupNode(params.GenesisHash)
		if !ok {
			return nil, errors.New("genesis is not found in the DAG after it was proccessed")
		}
	}

	// Save a reference to the genesis block.
	dag.genesis = genesis

	selectedTip := dag.selectedTip()
	log.Infof("DAG state (blue score %d, hash %s)",
		selectedTip.BlueScore, selectedTip.Hash)

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
func (dag *BlockDAG) selectedTip() *blocknode.Node {
	return dag.virtual.SelectedParent
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

	return selectedTip.Hash
}

// UTXOSet returns the DAG's UTXO set
func (dag *BlockDAG) UTXOSet() *utxo.FullUTXOSet {
	return dag.virtual.utxoSet
}

// CalcPastMedianTime returns the past median time of the DAG.
func (dag *BlockDAG) CalcPastMedianTime() mstime.Time {
	return dag.PastMedianTime(dag.virtual.SelectedParent)
}

// GetUTXOEntry returns the requested unspent transaction output. The returned
// instance must be treated as immutable since it is shared by all callers.
//
// This function is safe for concurrent access. However, the returned entry (if
// any) is NOT.
func (dag *BlockDAG) GetUTXOEntry(outpoint appmessage.Outpoint) (*utxo.Entry, bool) {
	dag.RLock()
	defer dag.RUnlock()
	return dag.virtual.utxoSet.Get(outpoint)
}

// BlueScoreByBlockHash returns the blue score of a block with the given hash.
func (dag *BlockDAG) BlueScoreByBlockHash(hash *daghash.Hash) (uint64, error) {
	node, ok := dag.Index.LookupNode(hash)
	if !ok {
		return 0, errors.Errorf("block %s is unknown", hash)
	}

	return node.BlueScore, nil
}

// BluesByBlockHash returns the blues of the block for the given hash.
func (dag *BlockDAG) BluesByBlockHash(hash *daghash.Hash) ([]*daghash.Hash, error) {
	node, ok := dag.Index.LookupNode(hash)
	if !ok {
		return nil, errors.Errorf("block %s is unknown", hash)
	}

	hashes := make([]*daghash.Hash, len(node.Blues))
	for i, blue := range node.Blues {
		hashes[i] = blue.Hash
	}

	return hashes, nil
}

// SelectedTipBlueScore returns the blue score of the selected tip.
func (dag *BlockDAG) SelectedTipBlueScore() uint64 {
	return dag.selectedTip().BlueScore
}

// VirtualBlueHashes returns the blue of the current virtual block
func (dag *BlockDAG) VirtualBlueHashes() []*daghash.Hash {
	dag.RLock()
	defer dag.RUnlock()
	hashes := make([]*daghash.Hash, len(dag.virtual.Blues))
	for i, blue := range dag.virtual.Blues {
		hashes[i] = blue.Hash
	}
	return hashes
}

// VirtualBlueScore returns the blue score of the current virtual block
func (dag *BlockDAG) VirtualBlueScore() uint64 {
	return dag.virtual.BlueScore
}

// BlockCount returns the number of blocks in the DAG
func (dag *BlockDAG) BlockCount() uint64 {
	return dag.blockCount
}

// TipHashes returns the hashes of the DAG's tips
func (dag *BlockDAG) TipHashes() []*daghash.Hash {
	return dag.tips.Hashes()
}

// ValidTipHashes returns the hashes of the DAG's valid tips
func (dag *BlockDAG) ValidTipHashes() []*daghash.Hash {
	return dag.validTips.Hashes()
}

// VirtualParentHashes returns the hashes of the virtual block's parents
func (dag *BlockDAG) VirtualParentHashes() []*daghash.Hash {
	return dag.virtual.Parents.Hashes()
}

// HeaderByHash returns the block header identified by the given hash or an
// error if it doesn't exist.
func (dag *BlockDAG) HeaderByHash(hash *daghash.Hash) (*appmessage.BlockHeader, error) {
	node, ok := dag.Index.LookupNode(hash)
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
	node, ok := dag.Index.LookupNode(hash)
	if !ok {
		str := fmt.Sprintf("block %s is not in the DAG", hash)
		return nil, ErrNotInDAG(str)

	}

	return node.Children.Hashes(), nil
}

// SelectedParentHash returns the selected parent hash of the block with the given hash in the
// DAG.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) SelectedParentHash(blockHash *daghash.Hash) (*daghash.Hash, error) {
	node, ok := dag.Index.LookupNode(blockHash)
	if !ok {
		str := fmt.Sprintf("block %s is not in the DAG", blockHash)
		return nil, ErrNotInDAG(str)

	}

	if node.SelectedParent == nil {
		return nil, nil
	}
	return node.SelectedParent.Hash, nil
}

// isInPast returns true if `node` is in the past of `other`
//
// Note: this method will return true if `node == other`
func (dag *BlockDAG) isInPast(node *blocknode.Node, other *blocknode.Node) (bool, error) {
	return dag.reachabilityTree.isInPast(node, other)
}

// isInPastOfAny returns true if `node` is in the past of any of `others`
//
// Note: this method will return true if `node` is in `others`
func (dag *BlockDAG) isInPastOfAny(node *blocknode.Node, others blocknode.Set) (bool, error) {
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
func (dag *BlockDAG) isAnyInPastOf(nodes blocknode.Set, other *blocknode.Node) (bool, error) {
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

// GetHeaders returns DAG headers ordered by blue score, starts from the given hash with the given direction.
func (dag *BlockDAG) GetHeaders(startHash *daghash.Hash, maxHeaders uint64,
	isAscending bool) ([]*appmessage.BlockHeader, error) {

	dag.RLock()
	defer dag.RUnlock()

	if isAscending {
		return dag.getHeadersAscending(startHash, maxHeaders)
	}

	return dag.getHeadersDescending(startHash, maxHeaders)
}

func (dag *BlockDAG) getHeadersDescending(highHash *daghash.Hash, maxHeaders uint64) ([]*appmessage.BlockHeader, error) {
	highNode := dag.virtual.Node
	if highHash != nil {
		var ok bool
		highNode, ok = dag.Index.LookupNode(highHash)
		if !ok {
			return nil, errors.Errorf("Couldn't find the start hash %s in the dag", highHash)
		}
	}
	headers := make([]*appmessage.BlockHeader, 0, maxHeaders)
	queue := blocknode.NewDownHeap()
	queue.PushSet(highNode.Parents)

	visited := blocknode.NewSet()
	for i := uint32(0); queue.Len() > 0 && uint64(len(headers)) < maxHeaders; i++ {
		current := queue.Pop()
		if !visited.Contains(current) {
			visited.Add(current)
			headers = append(headers, current.Header())
			queue.PushSet(current.Parents)
		}
	}
	return headers, nil
}

func (dag *BlockDAG) getHeadersAscending(lowHash *daghash.Hash, maxHeaders uint64) ([]*appmessage.BlockHeader, error) {
	lowNode := dag.genesis
	if lowHash != nil {
		var ok bool
		lowNode, ok = dag.Index.LookupNode(lowHash)
		if !ok {
			return nil, errors.Errorf("Couldn't find the start hash %s in the dag", lowHash)
		}
	}
	headers := make([]*appmessage.BlockHeader, 0, maxHeaders)
	queue := blocknode.NewUpHeap()
	queue.PushSet(lowNode.Children)

	visited := blocknode.NewSet()
	for i := uint32(0); queue.Len() > 0 && uint64(len(headers)) < maxHeaders; i++ {
		current := queue.Pop()
		if !visited.Contains(current) {
			visited.Add(current)
			headers = append(headers, current.Header())
			queue.PushSet(current.Children)
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
	for hash := range dag.Index.Index {
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
	return dag.Index.HaveBlock(hash)
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
	node, ok := dag.Index.LookupNode(hash)
	if !ok {
		return false
	}
	return dag.Index.BlockNodeStatus(node).KnownInvalid()
}

func (dag *BlockDAG) addTip(tip *blocknode.Node) (
	didVirtualParentsChange bool, chainUpdates *selectedParentChainUpdates, err error) {

	newTips := dag.tips.Clone()
	for parent := range tip.Parents {
		newTips.Remove(parent)
	}

	newTips.Add(tip)

	return dag.setTips(newTips)
}

func (dag *BlockDAG) setTips(newTips blocknode.Set) (
	didVirtualParentsChange bool, chainUpdates *selectedParentChainUpdates, err error) {

	didVirtualParentsChange, chainUpdates, err = dag.updateVirtualParents(newTips, dag.finalityPoint(dag.virtual.Node))
	if err != nil {
		return false, nil, err
	}

	dag.tips = newTips

	return didVirtualParentsChange, chainUpdates, nil
}

func (dag *BlockDAG) updateVirtualParents(newTips blocknode.Set, finalityPoint *blocknode.Node) (
	didVirtualParentsChange bool, chainUpdates *selectedParentChainUpdates, err error) {

	var newVirtualParents blocknode.Set
	// If only genesis is the newTips - we are still initializing the DAG and not all structures required
	// for calling dag.selectVirtualParents have been initialized yet.
	// Specifically - this function would be called with finalityPoint = dag.virtual.finalityPoint(), which has
	// not been initialized to anything real yet.
	//
	// Therefore, in this case - simply pick genesis as virtual's only parent
	if newTips.IsOnlyGenesis() {
		newVirtualParents = newTips
	} else {
		newVirtualParents, err = dag.selectVirtualParents(newTips, finalityPoint)
		if err != nil {
			return false, nil, err
		}
	}

	oldVirtualParents := dag.virtual.Parents
	didVirtualParentsChange = !oldVirtualParents.IsEqual(newVirtualParents)

	if !didVirtualParentsChange {
		return false, &selectedParentChainUpdates{}, nil
	}

	oldSelectedParent := dag.virtual.SelectedParent
	dag.virtual.Node, _ = dag.newBlockNode(nil, newVirtualParents)
	chainUpdates = dag.virtual.updateSelectedParentSet(oldSelectedParent)

	return didVirtualParentsChange, chainUpdates, nil
}

func (dag *BlockDAG) addValidTip(newValidTip *blocknode.Node) error {
	newValidTips := dag.validTips.Clone()
	for validTip := range dag.validTips {
		// We use isInPastOfAny on newValidTip.parents instead of
		// isInPast on newValidTip because newValidTip does not
		// necessarily have reachability data associated with it yet.
		isInPastOfNewValidTip, err := dag.isInPastOfAny(validTip, newValidTip.Parents)
		if err != nil {
			return err
		}
		if isInPastOfNewValidTip {
			newValidTips.Remove(validTip)
		}
	}

	newValidTips.Add(newValidTip)

	dag.validTips = newValidTips
	return nil
}

func (dag *BlockDAG) selectVirtualParents(tips blocknode.Set, finalityPoint *blocknode.Node) (blocknode.Set, error) {
	selected := blocknode.NewSet()

	candidatesHeap := blocknode.NewDownHeap()
	candidatesHeap.PushSet(tips)

	// If the first candidate has been disqualified from the chain or violates finality -
	// it cannot be virtual's parent, since it will make it virtual's selectedParent - disqualifying virtual itself.
	// Therefore, in such a case we remove it from the list of virtual parent candidates, and replace with
	// its parents that have no disqualified children
	disqualifiedCandidates := blocknode.NewSet()
	for {
		if candidatesHeap.Len() == 0 {
			return nil, errors.New("virtual has no valid parent candidates")
		}
		selectedParentCandidate := candidatesHeap.Pop()

		if dag.Index.BlockNodeStatus(selectedParentCandidate) == blocknode.StatusValid {
			isFinalityPointInSelectedParentChain, err := dag.isInSelectedParentChainOf(finalityPoint, selectedParentCandidate)
			if err != nil {
				return nil, err
			}

			if isFinalityPointInSelectedParentChain {
				selected.Add(selectedParentCandidate)
				break
			}
		}

		disqualifiedCandidates.Add(selectedParentCandidate)

		for parent := range selectedParentCandidate.Parents {
			if parent.Children.AreAllIn(disqualifiedCandidates) {
				candidatesHeap.Push(parent)
			}
		}
	}

	mergeSetSize := 1 // starts counting from 1 because selectedParent is already in the mergeSet

	for len(selected) < appmessage.MaxBlockParents && candidatesHeap.Len() > 0 {
		candidate := candidatesHeap.Pop()

		// check that the candidate doesn't increase the virtual's merge set over `mergeSetSizeLimit`
		mergeSetIncrease, err := dag.mergeSetIncrease(candidate, selected)
		if err != nil {
			return nil, err
		}

		if mergeSetSize+mergeSetIncrease > mergeSetSizeLimit {
			continue
		}

		selected.Add(candidate)
		mergeSetSize += mergeSetIncrease
	}

	tempVirtual, _ := dag.newBlockNode(nil, selected)

	boundedMergeBreakingParents, err := dag.boundedMergeBreakingParents(tempVirtual)
	if err != nil {
		return nil, err
	}
	selected = selected.Subtract(boundedMergeBreakingParents)

	return selected, nil
}

func (dag *BlockDAG) mergeSetIncrease(candidate *blocknode.Node, selected blocknode.Set) (int, error) {
	visited := blocknode.NewSet()
	queue := blocknode.NewDownHeap()
	queue.Push(candidate)
	mergeSetIncrease := 1 // starts with 1 for the candidate itself

	for queue.Len() > 0 {
		current := queue.Pop()
		isInPastOfSelected, err := dag.isInPastOfAny(current, selected)
		if err != nil {
			return 0, err
		}
		if isInPastOfSelected {
			continue
		}
		mergeSetIncrease++

		for parent := range current.Parents {
			if !visited.Contains(parent) {
				visited.Add(parent)
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
