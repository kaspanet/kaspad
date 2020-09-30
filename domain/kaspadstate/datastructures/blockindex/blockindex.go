package blockindex

import (
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

// BlockIndex ...
type BlockIndex struct {
}

// New ...
func New() *BlockIndex {
	return &BlockIndex{}
}

// Add ...
func (bi *BlockIndex) Add(dbTx *dbaccess.TxContext, blockHash *daghash.Hash) {

}

// Exists ...
func (bi *BlockIndex) Exists(dbContext dbaccess.Context, blockHash *daghash.Hash) bool {
	return false
}
