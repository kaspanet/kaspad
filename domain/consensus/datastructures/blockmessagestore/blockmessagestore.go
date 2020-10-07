package blockmessagestore

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/util/daghash"
)

// BlockMessageStore represents a store of MsgBlock
type BlockMessageStore struct {
}

// New instantiates a new BlockMessageStore
func New() *BlockMessageStore {
	return &BlockMessageStore{}
}

// Insert inserts the given msgBlock for the given blockHash
func (bms *BlockMessageStore) Insert(dbTx model.TxContextProxy, blockHash *daghash.Hash, msgBlock *appmessage.MsgBlock) {

}

// Get gets the msgBlock associated with the given blockHash
func (bms *BlockMessageStore) Get(dbContext model.ContextProxy, blockHash *daghash.Hash) *appmessage.MsgBlock {
	return nil
}
