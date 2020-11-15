package testapi

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// TestConsensus wraps the Consensus interface with some methods that are needed by tests only
type TestConsensus interface {
	externalapi.Consensus

	DatabaseContext() model.DBReader

	BuildBlockWithParents(parentHashes []*externalapi.DomainHash, coinbaseData *externalapi.DomainCoinbaseData,
		transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, error)

	// AddBlock builds a block with given information, solves it, and adds to the DAG.
	// Returns the hash of the added block
	AddBlock(parentHashes []*externalapi.DomainHash, coinbaseData *externalapi.DomainCoinbaseData,
		transactions []*externalapi.DomainTransaction) (*externalapi.DomainHash, error)

	SolveAndAddBlock(block *externalapi.DomainBlock) (*externalapi.DomainHash, error)

	AcceptanceDataStore() model.AcceptanceDataStore
	BlockHeaderStore() model.BlockHeaderStore
	BlockRelationStore() model.BlockRelationStore
	BlockStatusStore() model.BlockStatusStore
	BlockStore() model.BlockStore
	ConsensusStateStore() model.ConsensusStateStore
	GHOSTDAGDataStore() model.GHOSTDAGDataStore
	HeaderTipsStore() model.HeaderTipsStore
	MultisetStore() model.MultisetStore
	PruningStore() model.PruningStore
	ReachabilityDataStore() model.ReachabilityDataStore
	UTXODiffStore() model.UTXODiffStore

	BlockBuilder() model.BlockBuilder
	BlockProcessor() model.BlockProcessor
	BlockValidator() model.BlockValidator
	CoinbaseManager() model.CoinbaseManager
	ConsensusStateManager() model.ConsensusStateManager
	DAGTopologyManager() model.DAGTopologyManager
	DAGTraversalManager() model.DAGTraversalManager
	DifficultyManager() model.DifficultyManager
	GHOSTDAGManager() model.GHOSTDAGManager
	HeaderTipsManager() model.HeaderTipsManager
	MergeDepthManager() model.MergeDepthManager
	PastMedianTimeManager() model.PastMedianTimeManager
	PruningManager() model.PruningManager
	ReachabilityManager() model.TestReachabilityManager
	SyncManager() model.SyncManager
	TransactionValidator() model.TransactionValidator
}
