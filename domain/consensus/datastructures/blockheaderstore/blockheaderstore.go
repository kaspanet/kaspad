package blockheaderstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// blockHeaderStore represents a store of blocks
type blockHeaderStore struct {
}

// New instantiates a new BlockHeaderStore
func New() model.BlockHeaderStore {
	return &blockHeaderStore{}
}

// Stage stages the given block header for the given blockHash
func (bms *blockHeaderStore) Stage(blockHash *externalapi.DomainHash, block *externalapi.DomainBlockHeader) {
	panic("implement me")
}

func (bms *blockHeaderStore) IsStaged() bool {
	panic("implement me")
}

func (bms *blockHeaderStore) Discard() {
	panic("implement me")
}

func (bms *blockHeaderStore) Commit(dbTx model.DBTxProxy) error {
	panic("implement me")
}

// BlockHeader gets the block header associated with the given blockHash
func (bms *blockHeaderStore) BlockHeader(dbContext model.DBContextProxy, blockHash *externalapi.DomainHash) (*externalapi.DomainBlockHeader, error) {
	panic("implement me")
}

// HasBlock returns whether a block header with a given hash exists in the store.
func (bms *blockHeaderStore) HasBlockHeader(dbContext model.DBContextProxy, blockHash *externalapi.DomainHash) (bool, error) {
	panic("implement me")
}

// BlockHeaders gets the block headers associated with the given blockHashes
func (bms *blockHeaderStore) BlockHeaders(dbContext model.DBContextProxy, blockHashes []*externalapi.DomainHash) ([]*externalapi.DomainBlockHeader, error) {
	panic("implement me")
}

// Delete deletes the block associated with the given blockHash
func (bms *blockHeaderStore) Delete(dbTx model.DBTxProxy, blockHash *externalapi.DomainHash) error {
	panic("implement me")
}
