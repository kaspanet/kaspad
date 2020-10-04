package algorithms

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/kaspadstate/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
)

// BlockProcessor ...
type BlockProcessor interface {
	BuildBlock(transactionSelector model.TransactionSelector) *appmessage.MsgBlock
	ValidateAndInsertBlock(block *appmessage.MsgBlock) error
}

// BlockValidator ...
type BlockValidator interface {
	ValidateHeaderInIsolation(block *appmessage.MsgBlock) error
	ValidateHeaderInContext(block *appmessage.MsgBlock) error
	ValidateBodyInIsolation(block *appmessage.MsgBlock) error
	ValidateBodyInContext(block *appmessage.MsgBlock) error
}

// ConsensusStateManager ...
type ConsensusStateManager interface {
	UTXOByOutpoint(outpoint *appmessage.Outpoint) *model.UTXOEntry
	ValidateTransaction(transaction *util.Tx, utxoEntries []*model.UTXOEntry) error

	CalculateConsensusStateChanges(block *appmessage.MsgBlock) *model.ConsensusStateChanges
}

// DAGTopologyManager ...
type DAGTopologyManager interface {
	Parents(blockHash *daghash.Hash) []*daghash.Hash
	Children(blockHash *daghash.Hash) []*daghash.Hash
	IsParentOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool
	IsChildOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool
	IsAncestorOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool
	IsDescendantOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool
}

// DAGTraversalManager ...
type DAGTraversalManager interface {
	BlockAtDepth(uint64) *daghash.Hash
	SelectedParentIterator(highHash *daghash.Hash) model.SelectedParentIterator
}

// GHOSTDAGManager ...
type GHOSTDAGManager interface {
	GHOSTDAG(blockParents []*daghash.Hash) *model.BlockGHOSTDAGData
	BlockData(blockHash *daghash.Hash) *model.BlockGHOSTDAGData
}

// PruningManager ...
type PruningManager interface {
	FindNextPruningPoint(blockHash *daghash.Hash) (found bool, newPruningPoint *daghash.Hash, newPruningPointUTXOSet model.ReadOnlyUTXOSet)
	PruningPoint() *daghash.Hash
	SerializedUTXOSet() []byte
}

// ReachabilityTree ...
type ReachabilityTree interface {
	AddNode(dbTx *dbaccess.TxContext, blockHash *daghash.Hash)
	IsReachabilityAncestorOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool
	IsDAGAncestorOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool
}
