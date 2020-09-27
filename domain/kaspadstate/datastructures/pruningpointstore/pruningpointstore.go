package pruningpointstore

import (
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

type PruningPointStore interface {
	Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash)
	Get(dbContext dbaccess.Context) *daghash.Hash
}
