package model

import (
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

// ReachabilityDataStore represents a store of ReachabilityData
type ReachabilityDataStore interface {
	Insert(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, reachabilityData *ReachabilityData)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *ReachabilityData
}
