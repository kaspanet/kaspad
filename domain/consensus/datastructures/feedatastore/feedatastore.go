package feedatastore

import "github.com/kaspanet/kaspad/domain/consensus/model"

// feeDataStore represents a store of fee data
type feeDataStore struct {
}

// New instantiates a new feeDataStore
func New() model.FeeDataStore {
	return &feeDataStore{}
}

// Insert inserts the given fee for the given blockHash
func (ads *feeDataStore) Insert(dbTx model.DBTxProxy, blockHash *model.DomainHash, fee uint64) error {
	return nil
}

// Get gets the fee associated with the given blockHash
func (ads *feeDataStore) Get(dbContext model.DBContextProxy, blockHash *model.DomainHash) (uint64, error) {
	return 0, nil
}
