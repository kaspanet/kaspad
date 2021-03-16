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
	pruningManager      model.PruningManager

	ghostdagDataStore         model.GHOSTDAGDataStore
	blockStatusStore          model.BlockStatusStore
	blockHeaderStore          model.BlockHeaderStore
	blockStore                model.BlockStore
	pruningStore              model.PruningStore
	headersSelectedChainStore model.HeadersSelectedChainStore
}

// New instantiates a new SyncManager
func New(
	databaseContext model.DBReader,
	genesisBlockHash *externalapi.DomainHash,
	dagTraversalManager model.DAGTraversalManager,
	dagTopologyManager model.DAGTopologyManager,
	ghostdagManager model.GHOSTDAGManager,
	pruningManager model.PruningManager,

	ghostdagDataStore model.GHOSTDAGDataStore,
	blockStatusStore model.BlockStatusStore,
	blockHeaderStore model.BlockHeaderStore,
	blockStore model.BlockStore,
	pruningStore model.PruningStore,
	headersSelectedChainStore model.HeadersSelectedChainStore) model.SyncManager {

	return &syncManager{
		databaseContext:  databaseContext,
		genesisBlockHash: genesisBlockHash,

		dagTraversalManager:       dagTraversalManager,
		dagTopologyManager:        dagTopologyManager,
		ghostdagManager:           ghostdagManager,
		pruningManager:            pruningManager,
		headersSelectedChainStore: headersSelectedChainStore,

		ghostdagDataStore: ghostdagDataStore,
		blockStatusStore:  blockStatusStore,
		blockHeaderStore:  blockHeaderStore,
		blockStore:        blockStore,
		pruningStore:      pruningStore,
	}
}

func (sm *syncManager) GetHashesBetween(lowHash, highHash *externalapi.DomainHash,
	maxBlueScoreDifference uint64) (hashes []*externalapi.DomainHash, actualHighHash *externalapi.DomainHash, err error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "GetHashesBetween")
	defer onEnd()

	return sm.antiPastHashesBetween(lowHash, highHash, maxBlueScoreDifference)
}

func (sm *syncManager) GetMissingBlockBodyHashes(highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "GetMissingBlockBodyHashes")
	defer onEnd()

	return sm.missingBlockBodyHashes(highHash)
}

func (sm *syncManager) CreateBlockLocator(lowHash, highHash *externalapi.DomainHash, limit uint32) (externalapi.BlockLocator, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "CreateBlockLocator")
	defer onEnd()

	return sm.createBlockLocator(lowHash, highHash, limit)
}

func (sm *syncManager) CreateHeadersSelectedChainBlockLocator(lowHash,
	highHash *externalapi.DomainHash) (externalapi.BlockLocator, error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "CreateHeadersSelectedChainBlockLocator")
	defer onEnd()

	return sm.createHeadersSelectedChainBlockLocator(lowHash, highHash)
}

func (sm *syncManager) GetSyncInfo() (*externalapi.SyncInfo, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "GetSyncInfo")
	defer onEnd()

	return sm.syncInfo()
}
