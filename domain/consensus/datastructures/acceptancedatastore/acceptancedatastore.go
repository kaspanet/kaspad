package acceptancedatastore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// acceptanceDataStore represents a store of AcceptanceData
type acceptanceDataStore struct {
}

// New instantiates a new AcceptanceDataStore
func New() model.AcceptanceDataStore {
	return &acceptanceDataStore{}
}

// Stage stages the given acceptanceData for the given blockHash
func (ads *acceptanceDataStore) Stage(blockHash *externalapi.DomainHash, acceptanceData []*model.BlockAcceptanceData) {
	panic("implement me")
}

func (ads *acceptanceDataStore) IsStaged() bool {
	panic("implement me")
}

func (ads *acceptanceDataStore) Discard() {
	panic("implement me")
}

func (ads *acceptanceDataStore) Commit(dbTx model.DBTxProxy) error {
	panic("implement me")
}

// Get gets the acceptanceData associated with the given blockHash
func (ads *acceptanceDataStore) Get(dbContext model.DBContextProxy, blockHash *externalapi.DomainHash) ([]*model.BlockAcceptanceData, error) {
	return nil, nil
}

// Delete deletes the acceptanceData associated with the given blockHash
func (ads *acceptanceDataStore) Delete(dbTx model.DBTxProxy, blockHash *externalapi.DomainHash) error {
	return nil
}
