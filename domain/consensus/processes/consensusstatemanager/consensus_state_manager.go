package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// consensusStateManager manages the node's consensus state
type consensusStateManager struct {
	finalityDepth          uint64
	pruningDepth           uint64
	maxMassAcceptedByBlock uint64
	maxBlockParents        model.KType
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
	headerTipsStore       model.HeaderTipsStore

	blockStatusStore    model.BlockStatusStore
	ghostdagDataStore   model.GHOSTDAGDataStore
	consensusStateStore model.ConsensusStateStore
	multisetStore       model.MultisetStore
	blockStore          model.BlockStore
	utxoDiffStore       model.UTXODiffStore
	blockRelationStore  model.BlockRelationStore
	acceptanceDataStore model.AcceptanceDataStore
	blockHeaderStore    model.BlockHeaderStore

	stores []model.Store
}

// New instantiates a new ConsensusStateManager
func New(
	databaseContext model.DBManager,
	finalityDepth uint64,
	pruningDepth uint64,
	maxMassAcceptedByBlock uint64,
	maxBlockParents model.KType,
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

	blockStatusStore model.BlockStatusStore,
	ghostdagDataStore model.GHOSTDAGDataStore,
	consensusStateStore model.ConsensusStateStore,
	multisetStore model.MultisetStore,
	blockStore model.BlockStore,
	utxoDiffStore model.UTXODiffStore,
	blockRelationStore model.BlockRelationStore,
	acceptanceDataStore model.AcceptanceDataStore,
	blockHeaderStore model.BlockHeaderStore,
	headerTipsStore model.HeaderTipsStore) (model.ConsensusStateManager, error) {

	csm := &consensusStateManager{
		finalityDepth:          finalityDepth,
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

		multisetStore:       multisetStore,
		blockStore:          blockStore,
		blockStatusStore:    blockStatusStore,
		ghostdagDataStore:   ghostdagDataStore,
		consensusStateStore: consensusStateStore,
		utxoDiffStore:       utxoDiffStore,
		blockRelationStore:  blockRelationStore,
		acceptanceDataStore: acceptanceDataStore,
		blockHeaderStore:    blockHeaderStore,
		headerTipsStore:     headerTipsStore,

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
			headerTipsStore,
		},
	}

	return csm, nil
}
