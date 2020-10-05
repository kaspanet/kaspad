package blockindex

import (
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

// BlockIndex represents a store of known block hashes
type BlockIndex struct {
}

// New instantiates a new BlockIndex
func New() *BlockIndex {
	return &BlockIndex{}
}

// Insert ...
func (bi *BlockIndex) Insert(dbTx *dbaccess.TxContext, blockHash *daghash.Hash) {

}

// Exists ...
func (bi *BlockIndex) Exists(dbContext dbaccess.Context, blockHash *daghash.Hash) bool {
	return false
}
