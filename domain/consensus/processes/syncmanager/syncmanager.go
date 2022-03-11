package syncmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/pkg/errors"
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

	mergeSetSizeLimit uint64
}

// New instantiates a new SyncManager
func New(
	databaseContext model.DBReader,
	genesisBlockHash *externalapi.DomainHash,
	mergeSetSizeLimit uint64,
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

func (sm *syncManager) GetHashesBetween(stagingArea *model.StagingArea, lowHash, highHash *externalapi.DomainHash,
	maxBlocks uint64) (hashes []*externalapi.DomainHash, actualHighHash *externalapi.DomainHash, err error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "GetHashesBetween")
	defer onEnd()

	return sm.antiPastHashesBetween(stagingArea, lowHash, highHash, maxBlocks)
}

func (sm *syncManager) GetPastDiff(stagingArea *model.StagingArea, hasHash,
	requestedHash *externalapi.DomainHash, maxBlocks uint64) (hashes []*externalapi.DomainHash, err error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "GetPastDiff")
	defer onEnd()
	isHasAncestorOfRequested, err := sm.dagTopologyManager.IsAncestorOf(stagingArea, hasHash, requestedHash)
	if err != nil {
		return nil, err
	}
	if isHasAncestorOfRequested {
		return nil, errors.Errorf("expected block %s to be in anticone of %s",
			hasHash,
			requestedHash)
	}
	isRequestedAncestorOfHas, err := sm.dagTopologyManager.IsAncestorOf(stagingArea, requestedHash, hasHash)
	if err != nil {
		return nil, err
	}
	if isRequestedAncestorOfHas {
		return nil, errors.Errorf("expected block %s to be in anticone of %s",
			hasHash,
			requestedHash)
	}
	return sm.dagTraversalManager.AnticoneFromBlocks(stagingArea,
		[]*externalapi.DomainHash{requestedHash}, hasHash, maxBlocks)
}

func (sm *syncManager) GetMissingBlockBodyHashes(stagingArea *model.StagingArea, highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "GetMissingBlockBodyHashes")
	defer onEnd()

	return sm.missingBlockBodyHashes(stagingArea, highHash)
}

func (sm *syncManager) CreateBlockLocator(stagingArea *model.StagingArea,
	lowHash, highHash *externalapi.DomainHash, limit uint32) (externalapi.BlockLocator, error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "CreateBlockLocatorFromPruningPoint")
	defer onEnd()

	return sm.createBlockLocator(stagingArea, lowHash, highHash, limit)
}

func (sm *syncManager) CreateHeadersSelectedChainBlockLocator(stagingArea *model.StagingArea,
	lowHash, highHash *externalapi.DomainHash) (externalapi.BlockLocator, error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "CreateHeadersSelectedChainBlockLocator")
	defer onEnd()

	return sm.createHeadersSelectedChainBlockLocator(stagingArea, lowHash, highHash)
}

func (sm *syncManager) GetSyncInfo(stagingArea *model.StagingArea) (*externalapi.SyncInfo, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "GetSyncInfo")
	defer onEnd()

	return sm.syncInfo(stagingArea)
}
