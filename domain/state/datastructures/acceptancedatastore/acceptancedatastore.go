package acceptancedatastore

import (
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

type AcceptanceDataStore interface {
	Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, acceptanceData *AcceptanceData)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *AcceptanceData
}
