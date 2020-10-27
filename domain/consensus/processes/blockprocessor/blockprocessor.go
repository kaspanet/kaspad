package blockprocessor

import (
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/logger"
)

// blockProcessor is responsible for processing incoming blocks
// and creating blocks from the current state
type blockProcessor struct {
	dagParams       *dagconfig.Params
	databaseContext *database.DomainDBContext

	consensusStateManager model.ConsensusStateManager
	pruningManager        model.PruningManager
	blockValidator        model.BlockValidator
	dagTopologyManager    model.DAGTopologyManager
	reachabilityTree      model.ReachabilityTree
	difficultyManager     model.DifficultyManager
	ghostdagManager       model.GHOSTDAGManager
	pastMedianTimeManager model.PastMedianTimeManager
	coinbaseManager       model.CoinbaseManager

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

	stores []model.Store
}

// New instantiates a new BlockProcessor
func New(
	dagParams *dagconfig.Params,
	databaseContext *database.DomainDBContext,
	consensusStateManager model.ConsensusStateManager,
	pruningManager model.PruningManager,
	blockValidator model.BlockValidator,
	dagTopologyManager model.DAGTopologyManager,
	reachabilityTree model.ReachabilityTree,
	difficultyManager model.DifficultyManager,
	pastMedianTimeManager model.PastMedianTimeManager,
	ghostdagManager model.GHOSTDAGManager,
	coinbaseManager model.CoinbaseManager,
	acceptanceDataStore model.AcceptanceDataStore,
	blockStore model.BlockStore,
	blockStatusStore model.BlockStatusStore,
	blockRelationStore model.BlockRelationStore,
	multisetStore model.MultisetStore,
	ghostdagDataStore model.GHOSTDAGDataStore,
	consensusStateStore model.ConsensusStateStore,
	pruningStore model.PruningStore,
	reachabilityDataStore model.ReachabilityDataStore,
	utxoDiffStore model.UTXODiffStore) model.BlockProcessor {

	return &blockProcessor{
		dagParams:             dagParams,
		databaseContext:       databaseContext,
		pruningManager:        pruningManager,
		blockValidator:        blockValidator,
		dagTopologyManager:    dagTopologyManager,
		reachabilityTree:      reachabilityTree,
		difficultyManager:     difficultyManager,
		pastMedianTimeManager: pastMedianTimeManager,
		ghostdagManager:       ghostdagManager,
		coinbaseManager:       coinbaseManager,

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
		},
	}
}

// BuildBlock builds a block over the current state, with the given
// coinbaseData and the given transactions
func (bp *blockProcessor) BuildBlock(coinbaseData *externalapi.DomainCoinbaseData,
	transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "BuildBlock")
	defer onEnd()

	return bp.buildBlock(coinbaseData, transactions)
}

// ValidateAndInsertBlock validates the given block and, if valid, applies it
// to the current state
func (bp *blockProcessor) ValidateAndInsertBlock(block *externalapi.DomainBlock) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "ValidateAndInsertBlock")
	defer onEnd()

	return bp.validateAndInsertBlock(block)
}
