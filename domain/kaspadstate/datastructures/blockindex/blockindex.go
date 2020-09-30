package blockindex

import (
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

type BlockIndex struct {
}

func New() *BlockIndex {
	return &BlockIndex{}
}

func (bi *BlockIndex) Add(dbTx *dbaccess.TxContext, blockHash *daghash.Hash) {

}

func (bi *BlockIndex) Exists(dbContext dbaccess.Context, blockHash *daghash.Hash) bool {
	return false
}
