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
func (bms *BlockMessageStore) Insert(dbTx model.DBTxProxy, blockHash *model.DomainHash, msgBlock *model.DomainBlock) {

}

// Get gets the msgBlock associated with the given blockHash
func (bms *BlockMessageStore) Block(dbContext model.DBContextProxy, blockHash *model.DomainHash) *model.DomainBlock {
	return nil
}

func (bms *BlockMessageStore) Blocks(dbContext model.DBContextProxy, blockHashes []*model.DomainHash) []*model.DomainBlock {
	return nil
}

func (bms *BlockMessageStore) Delete(dbTx model.DBTxProxy, blockHash *model.DomainHash) {

}
