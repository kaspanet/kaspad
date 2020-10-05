package blockstatusstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

// BlockStatusStore ...
type BlockStatusStore struct {
}

// New ...
func New() *BlockStatusStore {
	return &BlockStatusStore{}
}

// Set ...
func (bss *BlockStatusStore) Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, blockStatus model.BlockStatus) {

}

// Get ...
func (bss *BlockStatusStore) Get(dbContext dbaccess.Context, blockHash *daghash.Hash) model.BlockStatus {
	return 0
}
