package syncmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"time"
)

type syncManager struct {
	databaseContext    model.DBReader
	genesisBlockHash   *externalapi.DomainHash
	targetTimePerBlock time.Duration

	dagTraversalManager   model.DAGTraversalManager
	dagTopologyManager    model.DAGTopologyManager
	ghostdagManager       model.GHOSTDAGManager
	consensusStateManager model.ConsensusStateManager

	ghostdagDataStore model.GHOSTDAGDataStore
	blockStatusStore  model.BlockStatusStore
	blockHeaderStore  model.BlockHeaderStore
	headerTipsStore   model.HeaderTipsStore
}

// New instantiates a new SyncManager
func New(
	databaseContext model.DBReader,
	genesisBlockHash *externalapi.DomainHash,
	targetTimePerBlock time.Duration,
	dagTraversalManager model.DAGTraversalManager,
	dagTopologyManager model.DAGTopologyManager,
	ghostdagManager model.GHOSTDAGManager,
	consensusStateManager model.ConsensusStateManager,

	ghostdagDataStore model.GHOSTDAGDataStore,
	blockStatusStore model.BlockStatusStore,
	blockHeaderStore model.BlockHeaderStore,
	headerTipsStore model.HeaderTipsStore) model.SyncManager {

	return &syncManager{
		databaseContext:    databaseContext,
		genesisBlockHash:   genesisBlockHash,
		targetTimePerBlock: targetTimePerBlock,

		dagTraversalManager:   dagTraversalManager,
		dagTopologyManager:    dagTopologyManager,
		ghostdagManager:       ghostdagManager,
		consensusStateManager: consensusStateManager,

		ghostdagDataStore: ghostdagDataStore,
		blockStatusStore:  blockStatusStore,
		blockHeaderStore:  blockHeaderStore,
		headerTipsStore:   headerTipsStore,
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

	return sm.isBlockInHeaderPruningPointFuture(blockHash)
}

func (sm *syncManager) CreateBlockLocator(lowHash, highHash *externalapi.DomainHash) (externalapi.BlockLocator, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "CreateBlockLocator")
	defer onEnd()

	return sm.createBlockLocator(lowHash, highHash)
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
