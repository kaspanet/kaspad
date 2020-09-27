package blockindex

import (
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

type BlockIndex interface {
	Add(dbTx *dbaccess.TxContext, blockHash *daghash.Hash)
	Exists(dbContext dbaccess.Context, blockHash *daghash.Hash) bool
}
