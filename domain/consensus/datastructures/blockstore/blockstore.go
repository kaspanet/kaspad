package blockstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// blockStore represents a store of blocks
type blockStore struct {
}

// New instantiates a new blockStore
func New() model.BlockStore {
	return &blockStore{}
}

// Insert inserts the given block for the given blockHash
func (bms *blockStore) Insert(dbTx model.DBTxProxy, blockHash *model.DomainHash, msgBlock *model.DomainBlock) {

}

// Block gets the block associated with the given blockHash
func (bms *blockStore) Block(dbContext model.DBContextProxy, blockHash *model.DomainHash) *model.DomainBlock {
	return nil
}

// Blocks gets the blocks associated with the given blockHashes
func (bms *blockStore) Blocks(dbContext model.DBContextProxy, blockHashes []*model.DomainHash) []*model.DomainBlock {
	return nil
}

// Delete deletes the block associated with the given blockHash
func (bms *blockStore) Delete(dbTx model.DBTxProxy, blockHash *model.DomainHash) {

}
