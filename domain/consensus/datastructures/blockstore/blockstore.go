package blockstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// blockStore represents a store of blocks
type blockStore struct {
}

// New instantiates a new BlockStore
func New() model.BlockStore {
	return &blockStore{}
}

// Insert inserts the given block for the given blockHash
func (bms *blockStore) Insert(dbTx model.DBTxProxy, blockHash *externalapi.DomainHash, msgBlock *externalapi.DomainBlock) error {
	return nil
}

// Block gets the block associated with the given blockHash
func (bms *blockStore) Block(dbContext model.DBContextProxy, blockHash *externalapi.DomainHash) (*externalapi.DomainBlock, error) {
	return nil, nil
}

// Blocks gets the blocks associated with the given blockHashes
func (bms *blockStore) Blocks(dbContext model.DBContextProxy, blockHashes []*externalapi.DomainHash) ([]*externalapi.DomainBlock, error) {
	return nil, nil
}

// Delete deletes the block associated with the given blockHash
func (bms *blockStore) Delete(dbTx model.DBTxProxy, blockHash *externalapi.DomainHash) error {
	return nil
}
