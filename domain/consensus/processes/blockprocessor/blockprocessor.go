package blockprocessor

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/logger"
)

// blockProcessor is responsible for processing incoming blocks
// and creating blocks from the current state
type blockProcessor struct {
	genesisHash     *externalapi.DomainHash
	databaseContext model.DBManager

	consensusStateManager model.ConsensusStateManager
	pruningManager        model.PruningManager
	blockValidator        model.BlockValidator
	dagTopologyManager    model.DAGTopologyManager
	reachabilityManager   model.ReachabilityManager
	difficultyManager     model.DifficultyManager
	ghostdagManager       model.GHOSTDAGManager
	pastMedianTimeManager model.PastMedianTimeManager
	coinbaseManager       model.CoinbaseManager
	headerTipsManager     model.HeaderTipsManager
	syncManager           model.SyncManager

	acceptanceDataStore   model.AcceptanceDataStore
	blockStore            model.BlockStore
	blockStatusStore      model.BlockStatusStore
	blockRelationStore    model.BlockRelationStore
	multisetStore         model.MultisetStore
	ghostdagDataStore     model.GHOSTDAGDataStore
	consensusStateStore   model.ConsensusStateStore
	pruningStore          model.PruningStore
	reachabilityDataStore model.ReachabilityDataStore
	utxoDiffStore         model.UTXODiffStore
	blockHeaderStore      model.BlockHeaderStore
	headerTipsStore       model.HeaderTipsStore
	finalityStore         model.FinalityStore

	stores []model.Store
}

// New instantiates a new BlockProcessor
func New(
	genesisHash *externalapi.DomainHash,
	databaseContext model.DBManager,
	consensusStateManager model.ConsensusStateManager,
	pruningManager model.PruningManager,
	blockValidator model.BlockValidator,
	dagTopologyManager model.DAGTopologyManager,
	reachabilityManager model.ReachabilityManager,
	difficultyManager model.DifficultyManager,
	pastMedianTimeManager model.PastMedianTimeManager,
	ghostdagManager model.GHOSTDAGManager,
	coinbaseManager model.CoinbaseManager,
	headerTipsManager model.HeaderTipsManager,
	syncManager model.SyncManager,

	acceptanceDataStore model.AcceptanceDataStore,
	blockStore model.BlockStore,
	blockStatusStore model.BlockStatusStore,
	blockRelationStore model.BlockRelationStore,
	multisetStore model.MultisetStore,
	ghostdagDataStore model.GHOSTDAGDataStore,
	consensusStateStore model.ConsensusStateStore,
	pruningStore model.PruningStore,
	reachabilityDataStore model.ReachabilityDataStore,
	utxoDiffStore model.UTXODiffStore,
	blockHeaderStore model.BlockHeaderStore,
	headerTipsStore model.HeaderTipsStore,
	finalityStore model.FinalityStore,
) model.BlockProcessor {

	return &blockProcessor{
		genesisHash:           genesisHash,
		databaseContext:       databaseContext,
		pruningManager:        pruningManager,
		blockValidator:        blockValidator,
		dagTopologyManager:    dagTopologyManager,
		reachabilityManager:   reachabilityManager,
		difficultyManager:     difficultyManager,
		pastMedianTimeManager: pastMedianTimeManager,
		ghostdagManager:       ghostdagManager,
		coinbaseManager:       coinbaseManager,
		headerTipsManager:     headerTipsManager,
		syncManager:           syncManager,

		consensusStateManager: consensusStateManager,
		acceptanceDataStore:   acceptanceDataStore,
		blockStore:            blockStore,
		blockStatusStore:      blockStatusStore,
		blockRelationStore:    blockRelationStore,
		multisetStore:         multisetStore,
		ghostdagDataStore:     ghostdagDataStore,
		consensusStateStore:   consensusStateStore,
		pruningStore:          pruningStore,
		reachabilityDataStore: reachabilityDataStore,
		utxoDiffStore:         utxoDiffStore,
		blockHeaderStore:      blockHeaderStore,
		headerTipsStore:       headerTipsStore,
		finalityStore:         finalityStore,

		stores: []model.Store{
			consensusStateStore,
			acceptanceDataStore,
			blockStore,
			blockStatusStore,
			blockRelationStore,
			multisetStore,
			ghostdagDataStore,
			consensusStateStore,
			pruningStore,
			reachabilityDataStore,
			utxoDiffStore,
			blockHeaderStore,
			headerTipsStore,
			finalityStore,
		},
	}
}

// ValidateAndInsertBlock validates the given block and, if valid, applies it
// to the current state
func (bp *blockProcessor) ValidateAndInsertBlock(block *externalapi.DomainBlock) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "ValidateAndInsertBlock")
	defer onEnd()

	return bp.validateAndInsertBlock(block)
}
