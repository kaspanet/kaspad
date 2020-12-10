package syncmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/logger"
)

type syncManager struct {
	databaseContext  model.DBReader
	genesisBlockHash *externalapi.DomainHash

	dagTraversalManager model.DAGTraversalManager
	dagTopologyManager  model.DAGTopologyManager
	ghostdagManager     model.GHOSTDAGManager

	ghostdagDataStore model.GHOSTDAGDataStore
	blockStatusStore  model.BlockStatusStore
	blockHeaderStore  model.BlockHeaderStore
	blockStore        model.BlockStore
	pruningStore      model.PruningStore
}

// New instantiates a new SyncManager
func New(
	databaseContext model.DBReader,
	genesisBlockHash *externalapi.DomainHash,
	dagTraversalManager model.DAGTraversalManager,
	dagTopologyManager model.DAGTopologyManager,
	ghostdagManager model.GHOSTDAGManager,

	ghostdagDataStore model.GHOSTDAGDataStore,
	blockStatusStore model.BlockStatusStore,
	blockHeaderStore model.BlockHeaderStore,
	blockStore model.BlockStore,
	pruningStore model.PruningStore) model.SyncManager {

	return &syncManager{
		databaseContext:  databaseContext,
		genesisBlockHash: genesisBlockHash,

		dagTraversalManager: dagTraversalManager,
		dagTopologyManager:  dagTopologyManager,
		ghostdagManager:     ghostdagManager,

		ghostdagDataStore: ghostdagDataStore,
		blockStatusStore:  blockStatusStore,
		blockHeaderStore:  blockHeaderStore,
		blockStore:        blockStore,
		pruningStore:      pruningStore,
	}
}

func (sm *syncManager) GetHashesBetween(lowHash, highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "GetHashesBetween")
	defer onEnd()

	return sm.antiPastHashesBetween(lowHash, highHash)
}

func (sm *syncManager) GetMissingBlockBodyHashes(highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "GetMissingBlockBodyHashes")
	defer onEnd()

	return sm.missingBlockBodyHashes(highHash)
}

func (sm *syncManager) IsBlockInHeaderPruningPointFuture(blockHash *externalapi.DomainHash) (bool, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "IsBlockInHeaderPruningPointFuture")
	defer onEnd()

	return sm.isBlockInPruningPointFuture(blockHash)
}

func (sm *syncManager) CreateBlockLocator(lowHash, highHash *externalapi.DomainHash, limit uint32) (externalapi.BlockLocator, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "CreateBlockLocator")
	defer onEnd()

	return sm.createBlockLocator(lowHash, highHash, limit)
}

func (sm *syncManager) FindNextBlockLocatorBoundaries(blockLocator externalapi.BlockLocator) (lowHash, highHash *externalapi.DomainHash, err error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "FindNextBlockLocatorBoundaries")
	defer onEnd()

	return sm.findNextBlockLocatorBoundaries(blockLocator)
}

func (sm *syncManager) GetSyncInfo() (*externalapi.SyncInfo, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "GetSyncInfo")
	defer onEnd()

	return sm.syncInfo()
}
