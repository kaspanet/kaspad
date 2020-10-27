package coinbasemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type coinbaseManager struct {
	ghostdagDataStore   model.GHOSTDAGDataStore
	acceptanceDataStore model.AcceptanceDataStore
}

// New instantiates a new CoinbaseManager
func New(
	ghostdagDataStore model.GHOSTDAGDataStore,
	acceptanceDataStore model.AcceptanceDataStore) model.CoinbaseManager {

	return &coinbaseManager{
		ghostdagDataStore:   ghostdagDataStore,
		acceptanceDataStore: acceptanceDataStore,
	}
}

func (c coinbaseManager) ExpectedCoinbaseTransaction(blockHash *externalapi.DomainHash,
	coinbaseData *externalapi.DomainCoinbaseData) (*externalapi.DomainTransaction, error) {

	panic("implement me")
}
