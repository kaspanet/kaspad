package ghostdagmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// ghostdagManager resolves and manages GHOSTDAG block data
type ghostdagManager struct {
	databaseContext    model.DBReader
	dagTopologyManager model.DAGTopologyManager
	ghostdagDataStore  model.GHOSTDAGDataStore
	headerStore        model.BlockHeaderStore
	k                  model.KType
}

// New instantiates a new GHOSTDAGManager
func New(
	databaseContext model.DBReader,
	dagTopologyManager model.DAGTopologyManager,
	ghostdagDataStore model.GHOSTDAGDataStore,
	headerStore model.BlockHeaderStore,
	k model.KType) model.GHOSTDAGManager {

	return &ghostdagManager{
		databaseContext:    databaseContext,
		dagTopologyManager: dagTopologyManager,
		ghostdagDataStore:  ghostdagDataStore,
		headerStore:        headerStore,
		k:                  k,
	}
}
