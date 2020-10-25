package acceptancemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// AcceptanceManager manages transaction acceptance
// and related data
type acceptanceManager struct {
	ghostdagDataStore   model.GHOSTDAGDataStore
	consensusStateStore model.ConsensusStateStore
	blockStore          model.BlockStore
	acceptanceDataStore model.AcceptanceDataStore
	multisetStore       model.MultisetStore
}

// New instantiates a new AcceptanceManager
func New(
	ghostdagDataStore model.GHOSTDAGDataStore,
	consensusStateStore model.ConsensusStateStore,
	blockStore model.BlockStore,
	acceptanceDataStore model.AcceptanceDataStore,
	multisetStore model.MultisetStore) model.AcceptanceManager {

	return &acceptanceManager{
		ghostdagDataStore:   ghostdagDataStore,
		consensusStateStore: consensusStateStore,
		blockStore:          blockStore,
		acceptanceDataStore: acceptanceDataStore,
		multisetStore:       multisetStore,
	}
}

func (a *acceptanceManager) CalculateAcceptanceData(blockHash *externalapi.DomainHash) error {
	panic("implement me")
}
