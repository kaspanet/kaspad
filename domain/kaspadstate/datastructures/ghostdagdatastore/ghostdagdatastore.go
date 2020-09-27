package ghostdagdatastore

import (
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

type GHOSTDAGDataStore interface {
	Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, blockGHOSTDAGData *BlockGHOSTDAGData)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *BlockGHOSTDAGData
}
