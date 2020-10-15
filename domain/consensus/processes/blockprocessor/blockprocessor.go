package blockprocessor

import (
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/dagconfig"
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
	acceptanceDataStore   model.AcceptanceDataStore
	blockMessageStore     model.BlockStore
	blockStatusStore      model.BlockStatusStore
	feeDataStore          model.FeeDataStore
}

// New instantiates a new blockProcessor
func New(
	dagParams *dagconfig.Params,
	databaseContext *database.DomainDBContext,
	consensusStateManager model.ConsensusStateManager,
	pruningManager model.PruningManager,
	blockValidator model.BlockValidator,
	dagTopologyManager model.DAGTopologyManager,
	reachabilityTree model.ReachabilityTree,
	acceptanceDataStore model.AcceptanceDataStore,
	blockMessageStore model.BlockStore,
	blockStatusStore model.BlockStatusStore,
	feeDataStore model.FeeDataStore) model.BlockProcessor {

	return &blockProcessor{
		dagParams:          dagParams,
		databaseContext:    databaseContext,
		pruningManager:     pruningManager,
		blockValidator:     blockValidator,
		dagTopologyManager: dagTopologyManager,
		reachabilityTree:   reachabilityTree,

		consensusStateManager: consensusStateManager,
		acceptanceDataStore:   acceptanceDataStore,
		blockMessageStore:     blockMessageStore,
		blockStatusStore:      blockStatusStore,
		feeDataStore:          feeDataStore,
	}
}

// BuildBlock builds a block over the current state, with the transactions
// selected by the given transactionSelector
func (bp *blockProcessor) BuildBlock(coinbaseScriptPublicKey []byte, coinbaseExtraData []byte,
	transactionSelector model.TransactionSelector) *model.DomainBlock {

	return nil
}

// ValidateAndInsertBlock validates the given block and, if valid, applies it
// to the current state
func (bp *blockProcessor) ValidateAndInsertBlock(block *model.DomainBlock) error {
	return nil
}
