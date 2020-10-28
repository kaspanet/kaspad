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

// Stage stages the given block for the given blockHash
func (bms *blockStore) Stage(blockHash *externalapi.DomainHash, block *externalapi.DomainBlock) {
	panic("implement me")
}

func (bms *blockStore) IsStaged() bool {
	panic("implement me")
}

func (bms *blockStore) Discard() {
	panic("implement me")
}

func (bms *blockStore) Commit(dbTx model.DBTransaction) error {
	panic("implement me")
}

// Block gets the block associated with the given blockHash
func (bms *blockStore) Block(dbContext model.DBReader, blockHash *externalapi.DomainHash) (*externalapi.DomainBlock, error) {
	panic("implement me")
}

// HasBlock returns whether a block with a given hash exists in the store.
func (bms *blockStore) HasBlock(dbContext model.DBReader, blockHash *externalapi.DomainHash) (bool, error) {
	panic("implement me")
}

// Blocks gets the blocks associated with the given blockHashes
func (bms *blockStore) Blocks(dbContext model.DBReader, blockHashes []*externalapi.DomainHash) ([]*externalapi.DomainBlock, error) {
	panic("implement me")
}

// Delete deletes the block associated with the given blockHash
func (bms *blockStore) Delete(dbTx model.DBTransaction, blockHash *externalapi.DomainHash) error {
	panic("implement me")
}
