package acceptancedatastore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// acceptanceDataStore represents a store of AcceptanceData
type acceptanceDataStore struct {
}

// New instantiates a new acceptanceDataStore
func New() model.AcceptanceDataStore {
	return &acceptanceDataStore{}
}

// Insert inserts the given acceptanceData for the given blockHash
func (ads *acceptanceDataStore) Insert(dbTx model.DBTxProxy, blockHash *model.DomainHash, acceptanceData *model.BlockAcceptanceData) {

}

// Get gets the acceptanceData associated with the given blockHash
func (ads *acceptanceDataStore) Get(dbContext model.DBContextProxy, blockHash *model.DomainHash) *model.BlockAcceptanceData {
	return nil
}

// Delete deletes the acceptanceData associated with the given blockHash
func (ads *acceptanceDataStore) Delete(dbTx model.DBTxProxy, blockHash *model.DomainHash) {

}
