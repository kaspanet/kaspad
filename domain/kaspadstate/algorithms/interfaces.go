package algorithms

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/kaspadstate/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
)

type BlockProcessor interface {
	BuildBlock(transactionSelector model.TransactionSelector) *appmessage.MsgBlock
	ValidateAndInsertBlock(block *appmessage.MsgBlock) error
}

type BlockValidator interface {
	ValidateHeaderInIsolation(block *appmessage.MsgBlock) error
	ValidateHeaderInContext(block *appmessage.MsgBlock) error
	ValidateBodyInIsolation(block *appmessage.MsgBlock) error
	ValidateBodyInContext(block *appmessage.MsgBlock) error
}

type ConsensusStateManager interface {
	UTXOByOutpoint(outpoint *appmessage.Outpoint) *model.UTXOEntry
	ValidateTransaction(transaction *util.Tx, utxoEntries []*model.UTXOEntry) error

	SerializedUTXOSet() []byte
	CalculateConsensusStateChanges(block *appmessage.MsgBlock) *model.ConsensusStateChanges
}

type DAGTopologyManager interface {
	Parents(blockHash *daghash.Hash) []*daghash.Hash
	Children(blockHash *daghash.Hash) []*daghash.Hash
	IsParentOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool
	IsChildOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool
	IsAncestorOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool
	IsDescendantOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool
}

type DAGTraversalManager interface {
	BlockAtDepth(uint64) *daghash.Hash
	SelectedParentIterator(highHash *daghash.Hash) model.SelectedParentIterator
}

type GHOSTDAGManager interface {
	GHOSTDAG(blockParents []*daghash.Hash) *model.BlockGHOSTDAGData
	BlockData(blockHash *daghash.Hash) *model.BlockGHOSTDAGData
}

type PruningManager interface {
	FindPruningPoint(blockHash *daghash.Hash) *daghash.Hash
}

type ReachabilityTree interface {
	AddNode(dbTx *dbaccess.TxContext, blockHash *daghash.Hash)
	IsInPastOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool
}
