package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// consensusStateManager manages the node's consensus state
type consensusStateManager struct {
	pruningDepth           uint64
	maxMassAcceptedByBlock uint64
	maxBlockParents        externalapi.KType
	mergeSetSizeLimit      uint64
	genesisHash            *externalapi.DomainHash
	databaseContext        model.DBManager

	ghostdagManager       model.GHOSTDAGManager
	dagTopologyManager    model.DAGTopologyManager
	dagTraversalManager   model.DAGTraversalManager
	pastMedianTimeManager model.PastMedianTimeManager
	transactionValidator  model.TransactionValidator
	blockValidator        model.BlockValidator
	reachabilityManager   model.ReachabilityManager
	coinbaseManager       model.CoinbaseManager
	mergeDepthManager     model.MergeDepthManager
	finalityManager       model.FinalityManager
	difficultyManager     model.DifficultyManager

	headersSelectedTipStore model.HeaderSelectedTipStore
	blockStatusStore        model.BlockStatusStore
	ghostdagDataStore       model.GHOSTDAGDataStore
	consensusStateStore     model.ConsensusStateStore
	multisetStore           model.MultisetStore
	blockStore              model.BlockStore
	utxoDiffStore           model.UTXODiffStore
	blockRelationStore      model.BlockRelationStore
	acceptanceDataStore     model.AcceptanceDataStore
	blockHeaderStore        model.BlockHeaderStore
	pruningStore            model.PruningStore
	daaBlocksStore          model.DAABlocksStore

	stores []model.Store
}

// New instantiates a new ConsensusStateManager
func New(
	databaseContext model.DBManager,
	pruningDepth uint64,
	maxMassAcceptedByBlock uint64,
	maxBlockParents externalapi.KType,
	mergeSetSizeLimit uint64,
	genesisHash *externalapi.DomainHash,

	ghostdagManager model.GHOSTDAGManager,
	dagTopologyManager model.DAGTopologyManager,
	dagTraversalManager model.DAGTraversalManager,
	pastMedianTimeManager model.PastMedianTimeManager,
	transactionValidator model.TransactionValidator,
	blockValidator model.BlockValidator,
	reachabilityManager model.ReachabilityManager,
	coinbaseManager model.CoinbaseManager,
	mergeDepthManager model.MergeDepthManager,
	finalityManager model.FinalityManager,
	difficultyManager model.DifficultyManager,

	blockStatusStore model.BlockStatusStore,
	ghostdagDataStore model.GHOSTDAGDataStore,
	consensusStateStore model.ConsensusStateStore,
	multisetStore model.MultisetStore,
	blockStore model.BlockStore,
	utxoDiffStore model.UTXODiffStore,
	blockRelationStore model.BlockRelationStore,
	acceptanceDataStore model.AcceptanceDataStore,
	blockHeaderStore model.BlockHeaderStore,
	headersSelectedTipStore model.HeaderSelectedTipStore,
	pruningStore model.PruningStore,
	daaBlocksStore model.DAABlocksStore) (model.ConsensusStateManager, error) {

	csm := &consensusStateManager{
		pruningDepth:           pruningDepth,
		maxMassAcceptedByBlock: maxMassAcceptedByBlock,
		maxBlockParents:        maxBlockParents,
		mergeSetSizeLimit:      mergeSetSizeLimit,
		genesisHash:            genesisHash,
		databaseContext:        databaseContext,

		ghostdagManager:       ghostdagManager,
		dagTopologyManager:    dagTopologyManager,
		dagTraversalManager:   dagTraversalManager,
		pastMedianTimeManager: pastMedianTimeManager,
		transactionValidator:  transactionValidator,
		blockValidator:        blockValidator,
		reachabilityManager:   reachabilityManager,
		coinbaseManager:       coinbaseManager,
		mergeDepthManager:     mergeDepthManager,
		finalityManager:       finalityManager,
		difficultyManager:     difficultyManager,

		multisetStore:           multisetStore,
		blockStore:              blockStore,
		blockStatusStore:        blockStatusStore,
		ghostdagDataStore:       ghostdagDataStore,
		consensusStateStore:     consensusStateStore,
		utxoDiffStore:           utxoDiffStore,
		blockRelationStore:      blockRelationStore,
		acceptanceDataStore:     acceptanceDataStore,
		blockHeaderStore:        blockHeaderStore,
		headersSelectedTipStore: headersSelectedTipStore,
		pruningStore:            pruningStore,
		daaBlocksStore:          daaBlocksStore,

		stores: []model.Store{
			consensusStateStore,
			acceptanceDataStore,
			blockStore,
			blockStatusStore,
			blockRelationStore,
			multisetStore,
			ghostdagDataStore,
			consensusStateStore,
			utxoDiffStore,
			blockHeaderStore,
			headersSelectedTipStore,
			pruningStore,
		},
	}

	return csm, nil
}
