package reachabilitydatastore

import (
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

type ReachabilityDataStore interface {
	Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, reachabilityData *ReachabilityData)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *ReachabilityData
}
