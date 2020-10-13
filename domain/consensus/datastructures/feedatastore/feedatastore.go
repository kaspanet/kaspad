package feedatastore

import "github.com/kaspanet/kaspad/domain/consensus/model"

// FeeDataStore represents a store of fee data
type FeeDataStore struct {
}

// New instantiates a new FeeDataStore
func New() *FeeDataStore {
	return &FeeDataStore{}
}

// Insert inserts the given fee for the given blockHash
func (ads *FeeDataStore) Insert(dbTx model.DBTxProxy, blockHash *model.DomainHash, fee uint64) {

}

// Get gets the fee associated with the given blockHash
func (ads *FeeDataStore) Get(dbContext model.DBContextProxy, blockHash *model.DomainHash) uint64 {
	return 0
}
