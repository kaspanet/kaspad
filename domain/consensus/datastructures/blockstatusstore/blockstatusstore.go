package blockstatusstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// blockStatusStore represents a store of BlockStatuses
type blockStatusStore struct {
}

// New instantiates a new blockStatusStore
func New() model.BlockStatusStore {
	return &blockStatusStore{}
}

// Insert inserts the given blockStatus for the given blockHash
func (bss *blockStatusStore) Insert(dbTx model.DBTxProxy, blockHash *model.DomainHash, blockStatus model.BlockStatus) error {
	return nil
}

// Get gets the blockStatus associated with the given blockHash
func (bss *blockStatusStore) Get(dbContext model.DBContextProxy, blockHash *model.DomainHash) (model.BlockStatus, error) {
	return 0, nil
}

// Exists returns true if the blockStatus for the given blockHash exists
func (bss *blockStatusStore) Exists(dbContext model.DBContextProxy, blockHash *model.DomainHash) (bool, error) {
	return false, nil
}
