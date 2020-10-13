package blockstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// BlockStore represents a store of blocks
type BlockStore struct {
}

// New instantiates a new BlockStore
func New() *BlockStore {
	return &BlockStore{}
}

// Insert inserts the given block for the given blockHash
func (bms *BlockStore) Insert(dbTx model.DBTxProxy, blockHash *model.DomainHash, msgBlock *model.DomainBlock) {

}

// Block gets the block associated with the given blockHash
func (bms *BlockStore) Block(dbContext model.DBContextProxy, blockHash *model.DomainHash) *model.DomainBlock {
	return nil
}

// Blocks gets the blocks associated with the given blockHashes
func (bms *BlockStore) Blocks(dbContext model.DBContextProxy, blockHashes []*model.DomainHash) []*model.DomainBlock {
	return nil
}

// Delete deletes the block associated with the given blockHash
func (bms *BlockStore) Delete(dbTx model.DBTxProxy, blockHash *model.DomainHash) {

}
