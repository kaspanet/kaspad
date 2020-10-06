package datastructures

import (
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

// AcceptanceDataStore represents a store of AcceptanceData
type AcceptanceDataStore interface {
	Insert(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, acceptanceData *model.BlockAcceptanceData)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *model.BlockAcceptanceData
}

// BlockIndex represents a store of known block hashes
type BlockIndex interface {
	Insert(dbTx *dbaccess.TxContext, blockHash *daghash.Hash)
	Exists(dbContext dbaccess.Context, blockHash *daghash.Hash) bool
}

// BlockMessageStore represents a store of MsgBlock
type BlockMessageStore interface {
	Insert(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, msgBlock *appmessage.MsgBlock)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *appmessage.MsgBlock
}

// BlockRelationStore represents a store of BlockRelations
type BlockRelationStore interface {
	Insert(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, blockRelationData *model.BlockRelations)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *model.BlockRelations
}

// BlockStatusStore represents a store of BlockStatuses
type BlockStatusStore interface {
	Insert(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, blockStatus model.BlockStatus)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) model.BlockStatus
}

// ConsensusStateStore represents a store for the current consensus state
type ConsensusStateStore interface {
	Update(dbTx *dbaccess.TxContext, utxoDiff *model.UTXODiff)
	UTXOByOutpoint(dbContext dbaccess.Context, outpoint *appmessage.Outpoint) *model.UTXOEntry
}

// GHOSTDAGDataStore represents a store of BlockGHOSTDAGData
type GHOSTDAGDataStore interface {
	Insert(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, blockGHOSTDAGData *model.BlockGHOSTDAGData)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *model.BlockGHOSTDAGData
}

// MultisetStore represents a store of Multisets
type MultisetStore interface {
	Insert(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, multiset *secp256k1.MultiSet)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *secp256k1.MultiSet
}

// PruningPointStore represents a store for the current pruning point
type PruningPointStore interface {
	Update(dbTx *dbaccess.TxContext, blockHash *daghash.Hash)
	Get(dbContext dbaccess.Context) *daghash.Hash
}

// ReachabilityDataStore represents a store of ReachabilityData
type ReachabilityDataStore interface {
	Insert(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, reachabilityData *model.ReachabilityData)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *model.ReachabilityData
}

// UTXODiffStore represents a store of UTXODiffs
type UTXODiffStore interface {
	Insert(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, utxoDiff *model.UTXODiff)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *model.UTXODiff
}
