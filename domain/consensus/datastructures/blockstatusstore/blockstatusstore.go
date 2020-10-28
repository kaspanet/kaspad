package blockstatusstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// blockStatusStore represents a store of BlockStatuses
type blockStatusStore struct {
}

// New instantiates a new BlockStatusStore
func New() model.BlockStatusStore {
	return &blockStatusStore{}
}

// Stage stages the given blockStatus for the given blockHash
func (bss *blockStatusStore) Stage(blockHash *externalapi.DomainHash, blockStatus model.BlockStatus) {
	panic("implement me")
}

func (bss *blockStatusStore) IsStaged() bool {
	panic("implement me")
}

func (bss *blockStatusStore) Discard() {
	panic("implement me")
}

func (bss *blockStatusStore) Commit(dbTx model.DBTransaction) error {
	panic("implement me")
}

// Get gets the blockStatus associated with the given blockHash
func (bss *blockStatusStore) Get(dbContext model.DBReader, blockHash *externalapi.DomainHash) (model.BlockStatus, error) {
	return 0, nil
}

// Exists returns true if the blockStatus for the given blockHash exists
func (bss *blockStatusStore) Exists(dbContext model.DBReader, blockHash *externalapi.DomainHash) (bool, error) {
	return false, nil
}
