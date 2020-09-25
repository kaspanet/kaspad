package blockstatusstore

import (
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

type BlockStatusStore interface {
	Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, blockStatus BlockStatus)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) BlockStatus
}
