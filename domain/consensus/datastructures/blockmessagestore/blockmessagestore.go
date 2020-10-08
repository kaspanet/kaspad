package blockmessagestore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// BlockMessageStore represents a store of MsgBlock
type BlockMessageStore struct {
}

// New instantiates a new BlockMessageStore
func New() *BlockMessageStore {
	return &BlockMessageStore{}
}

// Insert inserts the given msgBlock for the given blockHash
func (bms *BlockMessageStore) Insert(dbTx model.TxContextProxy, blockHash *model.DomainHash, msgBlock *model.DomainBlock) {

}

// Get gets the msgBlock associated with the given blockHash
func (bms *BlockMessageStore) Get(dbContext model.ContextProxy, blockHash *model.DomainHash) *model.DomainBlock {
	return nil
}
