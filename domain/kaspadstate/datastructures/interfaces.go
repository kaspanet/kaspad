package datastructures

import (
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/kaspadstate/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

type AcceptanceDataStore interface {
	Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, acceptanceData *model.AcceptanceData)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *model.AcceptanceData
}

type BlockIndex interface {
	Add(dbTx *dbaccess.TxContext, blockHash *daghash.Hash)
	Exists(dbContext dbaccess.Context, blockHash *daghash.Hash) bool
}

type BlockMessageStore interface {
	Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, msgBlock *appmessage.MsgBlock)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *appmessage.MsgBlock
}

type BlockRelationStore interface {
	Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, blockRelationData *model.BlockRelations)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *model.BlockRelations
}

type BlockStatusStore interface {
	Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, blockStatus model.BlockStatus)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) model.BlockStatus
}

type ConsensusStateStore interface {
	UpdateWithDiff(dbTx *dbaccess.TxContext, utxoDiff *model.UTXODiff)
	UTXOByOutpoint(dbContext dbaccess.Context, outpoint *appmessage.Outpoint) *model.UTXOEntry
}

type GHOSTDAGDataStore interface {
	Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, blockGHOSTDAGData *model.BlockGHOSTDAGData)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *model.BlockGHOSTDAGData
}

type MultisetStore interface {
	Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, multiset *secp256k1.MultiSet)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *secp256k1.MultiSet
}

type PruningPointStore interface {
	Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash)
	Get(dbContext dbaccess.Context) *daghash.Hash
}

type ReachabilityDataStore interface {
	Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, reachabilityData *model.ReachabilityData)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *model.ReachabilityData
}

type UTXODiffStore interface {
	Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, utxoDiff *model.UTXODiff)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *model.UTXODiff
}
