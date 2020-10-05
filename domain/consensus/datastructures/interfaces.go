package datastructures

import (
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

// AcceptanceDataStore ...
type AcceptanceDataStore interface {
	Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, acceptanceData *model.BlockAcceptanceData)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *model.BlockAcceptanceData
}

// BlockIndex ...
type BlockIndex interface {
	Add(dbTx *dbaccess.TxContext, blockHash *daghash.Hash)
	Exists(dbContext dbaccess.Context, blockHash *daghash.Hash) bool
}

// BlockMessageStore ...
type BlockMessageStore interface {
	Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, msgBlock *appmessage.MsgBlock)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *appmessage.MsgBlock
}

// BlockRelationStore ...
type BlockRelationStore interface {
	Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, blockRelationData *model.BlockRelations)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *model.BlockRelations
}

// BlockStatusStore ...
type BlockStatusStore interface {
	Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, blockStatus model.BlockStatus)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) model.BlockStatus
}

// ConsensusStateStore ...
type ConsensusStateStore interface {
	UpdateWithDiff(dbTx *dbaccess.TxContext, utxoDiff *model.UTXODiff)
	UTXOByOutpoint(dbContext dbaccess.Context, outpoint *appmessage.Outpoint) *model.UTXOEntry
}

// GHOSTDAGDataStore ...
type GHOSTDAGDataStore interface {
	Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, blockGHOSTDAGData *model.BlockGHOSTDAGData)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *model.BlockGHOSTDAGData
}

// MultisetStore ...
type MultisetStore interface {
	Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, multiset *secp256k1.MultiSet)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *secp256k1.MultiSet
}

// PruningPointStore ...
type PruningPointStore interface {
	Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash)
	Get(dbContext dbaccess.Context) *daghash.Hash
}

// ReachabilityDataStore ...
type ReachabilityDataStore interface {
	Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, reachabilityData *model.ReachabilityData)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *model.ReachabilityData
}

// UTXODiffStore ...
type UTXODiffStore interface {
	Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, utxoDiff *model.UTXODiff)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *model.UTXODiff
}
