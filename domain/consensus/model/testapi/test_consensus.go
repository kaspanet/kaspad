package testapi

import (
	"io"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
)

// MineJSONBlockType indicates which type of blocks MineJSON mines
type MineJSONBlockType int

const (
	// MineJSONBlockTypeUTXOValidBlock indicates for MineJSON to mine valid blocks.
	MineJSONBlockTypeUTXOValidBlock MineJSONBlockType = iota

	// MineJSONBlockTypeUTXOInvalidBlock indicates for MineJSON to mine UTXO invalid blocks.
	MineJSONBlockTypeUTXOInvalidBlock

	// MineJSONBlockTypeUTXOInvalidHeader indicates for MineJSON to mine UTXO invalid headers.
	MineJSONBlockTypeUTXOInvalidHeader
)

// TestConsensus wraps the Consensus interface with some methods that are needed by tests only
type TestConsensus interface {
	externalapi.Consensus

	DAGParams() *dagconfig.Params
	DatabaseContext() model.DBManager
	Database() database.Database

	BuildBlockWithParents(parentHashes []*externalapi.DomainHash, coinbaseData *externalapi.DomainCoinbaseData,
		transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, externalapi.UTXODiff, error)

	BuildHeaderWithParents(parentHashes []*externalapi.DomainHash) (externalapi.BlockHeader, error)

	BuildUTXOInvalidBlock(parentHashes []*externalapi.DomainHash) (*externalapi.DomainBlock, error)

	// AddBlock builds a block with given information, solves it, and adds to the DAG.
	// Returns the hash of the added block
	AddBlock(parentHashes []*externalapi.DomainHash, coinbaseData *externalapi.DomainCoinbaseData,
		transactions []*externalapi.DomainTransaction) (*externalapi.DomainHash, *externalapi.VirtualChangeSet, error)

	AddUTXOInvalidHeader(parentHashes []*externalapi.DomainHash) (*externalapi.DomainHash, *externalapi.VirtualChangeSet, error)

	AddUTXOInvalidBlock(parentHashes []*externalapi.DomainHash) (*externalapi.DomainHash,
		*externalapi.VirtualChangeSet, error)

	MineJSON(r io.Reader, blockType MineJSONBlockType) (tips []*externalapi.DomainHash, err error)
	ToJSON(w io.Writer) error

	RenderDAGToDot(filename string) error

	AcceptanceDataStore() model.AcceptanceDataStore
	BlockHeaderStore() model.BlockHeaderStore
	BlockRelationStore() model.BlockRelationStore
	BlockStatusStore() model.BlockStatusStore
	BlockStore() model.BlockStore
	ConsensusStateStore() model.ConsensusStateStore
	GHOSTDAGDataStore() model.GHOSTDAGDataStore
	GHOSTDAGDataStores() []model.GHOSTDAGDataStore
	HeaderTipsStore() model.HeaderSelectedTipStore
	MultisetStore() model.MultisetStore
	PruningStore() model.PruningStore
	ReachabilityDataStore() model.ReachabilityDataStore
	UTXODiffStore() model.UTXODiffStore
	HeadersSelectedChainStore() model.HeadersSelectedChainStore
	DAABlocksStore() model.DAABlocksStore

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
