package model

import (
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

// BlockIndex represents a store of known block hashes
type BlockIndex interface {
	Insert(dbTx *dbaccess.TxContext, blockHash *daghash.Hash)
	Exists(dbContext dbaccess.Context, blockHash *daghash.Hash) bool
}
