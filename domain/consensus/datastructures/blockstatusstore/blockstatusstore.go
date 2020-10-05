package blockstatusstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

// BlockStatusStore represents a store of BlockStatuses
type BlockStatusStore struct {
}

// New instantiates a new BlockStatusStore
func New() *BlockStatusStore {
	return &BlockStatusStore{}
}

// Insert ...
func (bss *BlockStatusStore) Insert(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, blockStatus model.BlockStatus) {

}

// Get ...
func (bss *BlockStatusStore) Get(dbContext dbaccess.Context, blockHash *daghash.Hash) model.BlockStatus {
	return 0
}
