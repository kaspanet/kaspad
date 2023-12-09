package ghostdagmanager

import (
	"github.com/zoomy-network/zoomyd/domain/consensus/model"
	"github.com/zoomy-network/zoomyd/domain/consensus/model/externalapi"
)

// ghostdagManager resolves and manages GHOSTDAG block data
type ghostdagManager struct {
	databaseContext    model.DBReader
	dagTopologyManager model.DAGTopologyManager
	ghostdagDataStore  model.GHOSTDAGDataStore
	headerStore        model.BlockHeaderStore

	k           externalapi.KType
	genesisHash *externalapi.DomainHash
}

// New instantiates a new GHOSTDAGManager
func New(
	databaseContext model.DBReader,
	dagTopologyManager model.DAGTopologyManager,
	ghostdagDataStore model.GHOSTDAGDataStore,
	headerStore model.BlockHeaderStore,
	k externalapi.KType,
	genesisHash *externalapi.DomainHash) model.GHOSTDAGManager {

	return &ghostdagManager{
		databaseContext:    databaseContext,
		dagTopologyManager: dagTopologyManager,
		ghostdagDataStore:  ghostdagDataStore,
		headerStore:        headerStore,
		k:                  k,
		genesisHash:        genesisHash,
	}
}
