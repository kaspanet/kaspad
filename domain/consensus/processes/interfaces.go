package processes

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
)

// BlockProcessor is responsible for processing incoming blocks
// and creating blocks from the current state
type BlockProcessor interface {
	BuildBlock(coinbaseScriptPublicKey []byte, coinbaseExtraData []byte, transactionSelector model.TransactionSelector) *appmessage.MsgBlock
	ValidateAndInsertBlock(block *appmessage.MsgBlock) error

	SetOnBlockAddedToDAGHandler(onBlockAddedToDAGHandler model.OnBlockAddedToDAGHandler)
	SetOnChainChangedHandler(onChainChangedHandler model.OnChainChangedHandler)
	SetOnFinalityConflictHandler(onFinalityConflictHandler model.OnFinalityConflictHandler)
}

// BlockValidator exposes a set of validation classes, after which
// it's possible to determine whether a block is valid
type BlockValidator interface {
	ValidateHeaderInIsolation(block *appmessage.MsgBlock) error
	ValidateBodyInIsolation(block *appmessage.MsgBlock) error
	ValidateHeaderInContext(block *appmessage.MsgBlock) error
	ValidateBodyInContext(block *appmessage.MsgBlock) error
	ValidateAgainstPastUTXO(block *appmessage.MsgBlock) error
	ValidateFinality(block *appmessage.MsgBlock) error
}

// ConsensusStateManager manages the node's consensus state
type ConsensusStateManager interface {
	UTXOByOutpoint(outpoint *appmessage.Outpoint) *model.UTXOEntry
	ValidateTransaction(transaction *util.Tx, utxoEntries []*model.UTXOEntry) error
	CalculateConsensusStateChanges(block *appmessage.MsgBlock) *model.ConsensusStateChanges
	ResolveFinalityConflict(blockHash *daghash.Hash)

	SetOnFinalityConflictResolvedHandler(onFinalityConflictResolvedHandler model.OnFinalityConflictResolvedHandler)
}

// DAGTopologyManager exposes methods for querying relationships
// between blocks in the DAG
type DAGTopologyManager interface {
	Parents(blockHash *daghash.Hash) []*daghash.Hash
	Children(blockHash *daghash.Hash) []*daghash.Hash
	IsParentOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool
	IsChildOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool
	IsAncestorOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool
	IsDescendantOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool
}

// DAGTraversalManager exposes methods for travering blocks
// in the DAG
type DAGTraversalManager interface {
	BlockAtDepth(highHash *daghash.Hash, depth uint64) *daghash.Hash
	SelectedParentIterator(highHash *daghash.Hash) model.SelectedParentIterator
}

// GHOSTDAGManager resolves and manages GHOSTDAG block data
type GHOSTDAGManager interface {
	GHOSTDAG(blockParents []*daghash.Hash) *model.BlockGHOSTDAGData
	BlockData(blockHash *daghash.Hash) *model.BlockGHOSTDAGData
}

// PruningManager resolves and manages the current pruning point
type PruningManager interface {
	FindNextPruningPoint(blockHash *daghash.Hash) (found bool, newPruningPoint *daghash.Hash, newPruningPointUTXOSet model.ReadOnlyUTXOSet)
	PruningPoint() *daghash.Hash
	SerializedUTXOSet() []byte
}

// ReachabilityTree maintains a structure that allows to answer
// reachability queries in sub-linear time
type ReachabilityTree interface {
	IsReachabilityTreeAncestorOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool
	IsDAGAncestorOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool
	ReachabilityChangeset(blockHash *daghash.Hash, blockGHOSTDAGData *model.BlockGHOSTDAGData) *model.ReachabilityChangeset
}
