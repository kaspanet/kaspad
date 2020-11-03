package syncmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/logger"
)

type syncManager struct {
	databaseContext model.DBReader

	dagTraversalManager model.DAGTraversalManager
	dagTopologyManager  model.DAGTopologyManager

	ghostdagDataStore model.GHOSTDAGDataStore
}

// New instantiates a new SyncManager
func New(
	databaseContext model.DBReader,
	dagTraversalManager model.DAGTraversalManager,
	dagTopologyManager model.DAGTopologyManager,
	ghostdagDataStore model.GHOSTDAGDataStore) model.SyncManager {

	return &syncManager{
		databaseContext: databaseContext,

		dagTraversalManager: dagTraversalManager,
		dagTopologyManager:  dagTopologyManager,

		ghostdagDataStore: ghostdagDataStore,
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

func (sm *syncManager) IsBlockHeaderInPruningPointFutureAndVirtualPast(blockHash *externalapi.DomainHash) (bool, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "IsBlockHeaderInPruningPointFutureAndVirtualPast")
	defer onEnd()

	return sm.isBlockHeaderInPruningPointFutureAndVirtualPast(blockHash)
}

func (sm *syncManager) CreateBlockLocator(lowHash, highHash *externalapi.DomainHash) (*externalapi.BlockLocator, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "CreateBlockLocator")
	defer onEnd()

	return sm.createBlockLocator(lowHash, highHash)
}

func (sm *syncManager) FindNextBlockLocatorBoundaries(blockLocator *externalapi.BlockLocator) (lowHash, highHash *externalapi.DomainHash, err error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "FindNextBlockLocatorBoundaries")
	defer onEnd()

	return sm.findNextBlockLocatorBoundaries(blockLocator)
}
