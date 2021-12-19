package blockprocessor

import (
	"time"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/processes/blockprocessor/blocklogger"
	"github.com/kaspanet/kaspad/infrastructure/logger"
)

// blockProcessor is responsible for processing incoming blocks
// and creating blocks from the current state
type blockProcessor struct {
	genesisHash        *externalapi.DomainHash
	targetTimePerBlock time.Duration
	databaseContext    model.DBManager
	blockLogger        *blocklogger.BlockLogger

	consensusStateManager model.ConsensusStateManager
	pruningManager        model.PruningManager
	blockValidator        model.BlockValidator
	dagTopologyManager    model.DAGTopologyManager
	reachabilityManagers  []model.ReachabilityManager
	difficultyManager     model.DifficultyManager
	pastMedianTimeManager model.PastMedianTimeManager
	coinbaseManager       model.CoinbaseManager
	headerTipsManager     model.HeadersSelectedTipManager
	syncManager           model.SyncManager

	acceptanceDataStore                 model.AcceptanceDataStore
	blockStore                          model.BlockStore
	blockStatusStore                    model.BlockStatusStore
	blockRelationStore                  model.BlockRelationStore
	multisetStore                       model.MultisetStore
	ghostdagDataStore                   model.GHOSTDAGDataStore
	consensusStateStore                 model.ConsensusStateStore
	pruningStore                        model.PruningStore
	reachabilityDataStore               model.ReachabilityDataStore
	utxoDiffStore                       model.UTXODiffStore
	blockHeaderStore                    model.BlockHeaderStore
	headersSelectedTipStore             model.HeaderSelectedTipStore
	finalityStore                       model.FinalityStore
	headersSelectedChainStore           model.HeadersSelectedChainStore
	daaBlocksStore                      model.DAABlocksStore
	blocksWithTrustedDataDAAWindowStore model.BlocksWithTrustedDataDAAWindowStore

	stores []model.Store
}

// New instantiates a new BlockProcessor
func New(
	genesisHash *externalapi.DomainHash,
	targetTimePerBlock time.Duration,
	databaseContext model.DBManager,

	consensusStateManager model.ConsensusStateManager,
	pruningManager model.PruningManager,
	blockValidator model.BlockValidator,
	dagTopologyManager model.DAGTopologyManager,
	reachabilityManagers []model.ReachabilityManager,
	difficultyManager model.DifficultyManager,
	pastMedianTimeManager model.PastMedianTimeManager,
	coinbaseManager model.CoinbaseManager,
	headerTipsManager model.HeadersSelectedTipManager,
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
	headersSelectedTipStore model.HeaderSelectedTipStore,
	finalityStore model.FinalityStore,
	headersSelectedChainStore model.HeadersSelectedChainStore,
	daaBlocksStore model.DAABlocksStore,
	blocksWithTrustedDataDAAWindowStore model.BlocksWithTrustedDataDAAWindowStore,
) model.BlockProcessor {

	return &blockProcessor{
		genesisHash:           genesisHash,
		targetTimePerBlock:    targetTimePerBlock,
		databaseContext:       databaseContext,
		blockLogger:           blocklogger.NewBlockLogger(),
		pruningManager:        pruningManager,
		blockValidator:        blockValidator,
		dagTopologyManager:    dagTopologyManager,
		reachabilityManagers:  reachabilityManagers,
		difficultyManager:     difficultyManager,
		pastMedianTimeManager: pastMedianTimeManager,
		coinbaseManager:       coinbaseManager,
		headerTipsManager:     headerTipsManager,
		syncManager:           syncManager,

		consensusStateManager:               consensusStateManager,
		acceptanceDataStore:                 acceptanceDataStore,
		blockStore:                          blockStore,
		blockStatusStore:                    blockStatusStore,
		blockRelationStore:                  blockRelationStore,
		multisetStore:                       multisetStore,
		ghostdagDataStore:                   ghostdagDataStore,
		consensusStateStore:                 consensusStateStore,
		pruningStore:                        pruningStore,
		reachabilityDataStore:               reachabilityDataStore,
		utxoDiffStore:                       utxoDiffStore,
		blockHeaderStore:                    blockHeaderStore,
		headersSelectedTipStore:             headersSelectedTipStore,
		finalityStore:                       finalityStore,
		headersSelectedChainStore:           headersSelectedChainStore,
		daaBlocksStore:                      daaBlocksStore,
		blocksWithTrustedDataDAAWindowStore: blocksWithTrustedDataDAAWindowStore,

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
			headersSelectedTipStore,
			finalityStore,
			headersSelectedChainStore,
			daaBlocksStore,
			blocksWithTrustedDataDAAWindowStore,
		},
	}
}

// ValidateAndInsertBlock validates the given block and, if valid, applies it
// to the current state
func (bp *blockProcessor) ValidateAndInsertBlock(block *externalapi.DomainBlock, shouldValidateAgainstUTXO bool) (*externalapi.VirtualChangeSet, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "ValidateAndInsertBlock")
	defer onEnd()

	stagingArea := model.NewStagingArea()
	return bp.validateAndInsertBlock(stagingArea, block, false, shouldValidateAgainstUTXO, false)
}

func (bp *blockProcessor) ValidateAndInsertImportedPruningPoint(newPruningPoint *externalapi.DomainHash) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "ValidateAndInsertImportedPruningPoint")
	defer onEnd()

	stagingArea := model.NewStagingArea()
	return bp.validateAndInsertImportedPruningPoint(stagingArea, newPruningPoint)
}

func (bp *blockProcessor) ValidateAndInsertBlockWithTrustedData(block *externalapi.BlockWithTrustedData, shouldValidateAgainstUTXO bool) (*externalapi.VirtualChangeSet, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "ValidateAndInsertBlockWithTrustedData")
	defer onEnd()

	stagingArea := model.NewStagingArea()

	return bp.validateAndInsertBlockWithTrustedData(stagingArea, block, shouldValidateAgainstUTXO)
}
