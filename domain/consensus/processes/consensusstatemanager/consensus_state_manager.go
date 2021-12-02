package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// consensusStateManager manages the node's consensus state
type consensusStateManager struct {
	maxBlockParents   externalapi.KType
	mergeSetSizeLimit uint64
	genesisHash       *externalapi.DomainHash
	databaseContext   model.DBManager

	ghostdagManager       model.GHOSTDAGManager
	dagTopologyManager    model.DAGTopologyManager
	dagTraversalManager   model.DAGTraversalManager
	pastMedianTimeManager model.PastMedianTimeManager
	transactionValidator  model.TransactionValidator
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

	onResolveVirtualHandler func(bir *externalapi.BlockInsertionResult) error
}

// New instantiates a new ConsensusStateManager
func New(
	databaseContext model.DBManager,
	maxBlockParents externalapi.KType,
	mergeSetSizeLimit uint64,
	genesisHash *externalapi.DomainHash,

	ghostdagManager model.GHOSTDAGManager,
	dagTopologyManager model.DAGTopologyManager,
	dagTraversalManager model.DAGTraversalManager,
	pastMedianTimeManager model.PastMedianTimeManager,
	transactionValidator model.TransactionValidator,
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
		maxBlockParents:   maxBlockParents,
		mergeSetSizeLimit: mergeSetSizeLimit,
		genesisHash:       genesisHash,
		databaseContext:   databaseContext,

		ghostdagManager:       ghostdagManager,
		dagTopologyManager:    dagTopologyManager,
		dagTraversalManager:   dagTraversalManager,
		pastMedianTimeManager: pastMedianTimeManager,
		transactionValidator:  transactionValidator,
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

// SetOnResolveVirtualHandler sets the onResolveVirtualHandler handler
func (csm *consensusStateManager) SetOnResolveVirtualHandler(onResolveVirtualHandler func(*externalapi.BlockInsertionResult) error) {
	csm.onResolveVirtualHandler = onResolveVirtualHandler
}
