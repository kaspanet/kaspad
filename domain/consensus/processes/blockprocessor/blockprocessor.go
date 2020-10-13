package blockprocessor

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
)

// BlockProcessor is responsible for processing incoming blocks
// and creating blocks from the current state
type BlockProcessor struct {
	dagParams       *dagconfig.Params
	databaseContext *dbaccess.DatabaseContext

	consensusStateManager model.ConsensusStateManager
	pruningManager        model.PruningManager
	blockValidator        model.BlockValidator
	dagTopologyManager    model.DAGTopologyManager
	reachabilityTree      model.ReachabilityTree
	acceptanceDataStore   model.AcceptanceDataStore
	blockMessageStore     model.BlockMessageStore
	blockStatusStore      model.BlockStatusStore
}

// New instantiates a new BlockProcessor
func New(
	dagParams *dagconfig.Params,
	databaseContext *dbaccess.DatabaseContext,
	consensusStateManager model.ConsensusStateManager,
	pruningManager model.PruningManager,
	blockValidator model.BlockValidator,
	dagTopologyManager model.DAGTopologyManager,
	reachabilityTree model.ReachabilityTree,
	acceptanceDataStore model.AcceptanceDataStore,
	blockMessageStore model.BlockMessageStore,
	blockStatusStore model.BlockStatusStore) *BlockProcessor {

	return &BlockProcessor{
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
	}
}

// BuildBlock builds a block over the current state, with the transactions
// selected by the given transactionSelector
func (bp *BlockProcessor) BuildBlock(coinbaseScriptPublicKey []byte, coinbaseExtraData []byte,
	transactionSelector model.TransactionSelector) *model.DomainBlock {

	return nil
}

// ValidateAndInsertBlock validates the given block and, if valid, applies it
// to the current state
func (bp *BlockProcessor) ValidateAndInsertBlock(block *model.DomainBlock) error {
	return nil
}
