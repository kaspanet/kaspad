package model

import (
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

// BlockStatusStore represents a store of BlockStatuses
type BlockStatusStore interface {
	Insert(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, blockStatus BlockStatus)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) BlockStatus
}
