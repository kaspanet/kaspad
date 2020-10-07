package model

import (
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

// PruningPointStore represents a store for the current pruning point
type PruningPointStore interface {
	Update(dbTx *dbaccess.TxContext, blockHash *daghash.Hash)
	Get(dbContext dbaccess.Context) *daghash.Hash
}
