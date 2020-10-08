package acceptancedatastore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// AcceptanceDataStore represents a store of AcceptanceData
type AcceptanceDataStore struct {
}

// New instantiates a new AcceptanceDataStore
func New() *AcceptanceDataStore {
	return &AcceptanceDataStore{}
}

// Insert inserts the given acceptanceData for the given blockHash
func (ads *AcceptanceDataStore) Insert(dbTx model.TxContextProxy, blockHash *model.DomainHash, acceptanceData *model.BlockAcceptanceData) {

}

// Get gets the acceptanceData associated with the given blockHash
func (ads *AcceptanceDataStore) Get(dbContext model.ContextProxy, blockHash *model.DomainHash) *model.BlockAcceptanceData {
	return nil
}
