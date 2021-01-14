package testapi

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
)

// TestConsensus wraps the Consensus interface with some methods that are needed by tests only
type TestConsensus interface {
	externalapi.Consensus

	DAGParams() *dagconfig.Params
	DatabaseContext() model.DBManager
	Database() database.Database

	BuildBlockWithParents(parentHashes []*externalapi.DomainHash, coinbaseData *externalapi.DomainCoinbaseData,
		transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, model.UTXODiff, error)

	BuildHeaderWithParents(parentHashes []*externalapi.DomainHash) (externalapi.BlockHeader, error)

	BuildUTXOInvalidBlock(parentHashes []*externalapi.DomainHash) (*externalapi.DomainBlock, error)

	// AddBlock builds a block with given information, solves it, and adds to the DAG.
	// Returns the hash of the added block
	AddBlock(parentHashes []*externalapi.DomainHash, coinbaseData *externalapi.DomainCoinbaseData,
		transactions []*externalapi.DomainTransaction) (*externalapi.DomainHash, *externalapi.BlockInsertionResult, error)

	AddUTXOInvalidHeader(parentHashes []*externalapi.DomainHash) (*externalapi.DomainHash, *externalapi.BlockInsertionResult, error)

	AddUTXOInvalidBlock(parentHashes []*externalapi.DomainHash) (*externalapi.DomainHash,
		*externalapi.BlockInsertionResult, error)

	DiscardAllStores()

	AcceptanceDataStore() model.AcceptanceDataStore
	BlockHeaderStore() model.BlockHeaderStore
	BlockRelationStore() model.BlockRelationStore
	BlockStatusStore() model.BlockStatusStore
	BlockStore() model.BlockStore
	ConsensusStateStore() model.ConsensusStateStore
	GHOSTDAGDataStore() model.GHOSTDAGDataStore
	HeaderTipsStore() model.HeaderSelectedTipStore
	MultisetStore() model.MultisetStore
	PruningStore() model.PruningStore
	ReachabilityDataStore() model.ReachabilityDataStore
	UTXODiffStore() model.UTXODiffStore
	HeadersSelectedChainStore() model.HeadersSelectedChainStore

	BlockBuilder() TestBlockBuilder
	BlockProcessor() model.BlockProcessor
	BlockValidator() model.BlockValidator
	CoinbaseManager() model.CoinbaseManager
	ConsensusStateManager() TestConsensusStateManager
	FinalityManager() model.FinalityManager
	DAGTopologyManager() model.DAGTopologyManager
	DAGTraversalManager() model.DAGTraversalManager
	DifficultyManager() model.DifficultyManager
	GHOSTDAGManager() model.GHOSTDAGManager
	HeaderTipsManager() model.HeadersSelectedTipManager
	MergeDepthManager() model.MergeDepthManager
	PastMedianTimeManager() model.PastMedianTimeManager
	PruningManager() model.PruningManager
	ReachabilityManager() TestReachabilityManager
	SyncManager() model.SyncManager
	TransactionValidator() TestTransactionValidator
}
