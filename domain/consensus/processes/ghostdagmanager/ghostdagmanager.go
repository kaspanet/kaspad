package ghostdagmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// ghostdagManager resolves and manages GHOSTDAG block data
type ghostdagManager struct {
	databaseContext                     model.DBReader
	dagTopologyManager                  model.DAGTopologyManager
	ghostdagDataStore                   model.GHOSTDAGDataStore
	blocksWithMetaDataGHOSTDAGDataStore model.GHOSTDAGDataStore
	headerStore                         model.BlockHeaderStore

	k           externalapi.KType
	genesisHash *externalapi.DomainHash
}

// New instantiates a new GHOSTDAGManager
func New(
	databaseContext model.DBReader,
	dagTopologyManager model.DAGTopologyManager,
	ghostdagDataStore model.GHOSTDAGDataStore,
	headerStore model.BlockHeaderStore,
	blocksWithMetaDataGHOSTDAGDataStore model.GHOSTDAGDataStore,
	k externalapi.KType,
	genesisHash *externalapi.DomainHash) model.GHOSTDAGManager {

	return &ghostdagManager{
		databaseContext:                     databaseContext,
		dagTopologyManager:                  dagTopologyManager,
		ghostdagDataStore:                   ghostdagDataStore,
		headerStore:                         headerStore,
		blocksWithMetaDataGHOSTDAGDataStore: blocksWithMetaDataGHOSTDAGDataStore,
		k:                                   k,
		genesisHash:                         genesisHash,
	}
}
