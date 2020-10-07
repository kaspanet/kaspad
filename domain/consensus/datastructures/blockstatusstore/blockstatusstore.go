package blockstatusstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/util/daghash"
)

// BlockStatusStore represents a store of BlockStatuses
type BlockStatusStore struct {
}

// New instantiates a new BlockStatusStore
func New() *BlockStatusStore {
	return &BlockStatusStore{}
}

// Insert inserts the given blockStatus for the given blockHash
func (bss *BlockStatusStore) Insert(dbTx model.TxContextProxy, blockHash *daghash.Hash, blockStatus model.BlockStatus) {

}

// Get gets the blockStatus associated with the given blockHash
func (bss *BlockStatusStore) Get(dbContext model.ContextProxy, blockHash *daghash.Hash) model.BlockStatus {
	return 0
}
