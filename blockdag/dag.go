// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"fmt"
	"github.com/kaspanet/kaspad/network/domainmessage"
	"github.com/kaspanet/kaspad/util/mstime"
	"sync"

	"github.com/kaspanet/kaspad/infrastructure/dbaccess"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/util/subnetworkid"

	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/txscript"
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
func (dag *BlockDAG) SelectedTipHeader() *domainmessage.BlockHeader {
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
func (dag *BlockDAG) GetUTXOEntry(outpoint domainmessage.Outpoint) (*UTXOEntry, bool) {
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

// HeaderByHash returns the block header identified by the given hash or an
// error if it doesn't exist.
func (dag *BlockDAG) HeaderByHash(hash *daghash.Hash) (*domainmessage.BlockHeader, error) {
	node, ok := dag.index.LookupNode(hash)
	if !ok {
		err := errors.Errorf("block %s is not known", hash)
		return &domainmessage.BlockHeader{}, err
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

func (dag *BlockDAG) isInPast(this *blockNode, other *blockNode) (bool, error) {
	return dag.reachabilityTree.isInPast(this, other)
}

// GetTopHeaders returns the top domainmessage.MaxBlockHeadersPerMsg block headers ordered by blue score.
func (dag *BlockDAG) GetTopHeaders(highHash *daghash.Hash, maxHeaders uint64) ([]*domainmessage.BlockHeader, error) {
	highNode := &dag.virtual.blockNode
	if highHash != nil {
		var ok bool
		highNode, ok = dag.index.LookupNode(highHash)
		if !ok {
			return nil, errors.Errorf("Couldn't find the high hash %s in the dag", highHash)
		}
	}
	headers := make([]*domainmessage.BlockHeader, 0, highNode.blueScore)
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
	return dag.index.NodeStatus(node).KnownInvalid()
}
