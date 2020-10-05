package blockprocessor

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/kaspadstate/algorithms"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures"
	"github.com/kaspanet/kaspad/domain/kaspadstate/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
)

// BlockProcessor ...
type BlockProcessor struct {
	dagParams       *dagconfig.Params
	databaseContext *dbaccess.DatabaseContext

	consensusStateManager algorithms.ConsensusStateManager
	pruningManager        algorithms.PruningManager
	blockValidator        algorithms.BlockValidator
	dagTopologyManager    algorithms.DAGTopologyManager
	reachabilityTree      algorithms.ReachabilityTree
	acceptanceDataStore   datastructures.AcceptanceDataStore
	blockIndex            datastructures.BlockIndex
	blockMessageStore     datastructures.BlockMessageStore
	blockStatusStore      datastructures.BlockStatusStore
}

// New ...
func New(
	dagParams *dagconfig.Params,
	databaseContext *dbaccess.DatabaseContext,
	consensusStateManager algorithms.ConsensusStateManager,
	pruningManager algorithms.PruningManager,
	blockValidator algorithms.BlockValidator,
	dagTopologyManager algorithms.DAGTopologyManager,
	reachabilityTree algorithms.ReachabilityTree,
	acceptanceDataStore datastructures.AcceptanceDataStore,
	blockIndex datastructures.BlockIndex,
	blockMessageStore datastructures.BlockMessageStore,
	blockStatusStore datastructures.BlockStatusStore) *BlockProcessor {

	return &BlockProcessor{
		dagParams:          dagParams,
		databaseContext:    databaseContext,
		pruningManager:     pruningManager,
		blockValidator:     blockValidator,
		dagTopologyManager: dagTopologyManager,
		reachabilityTree:   reachabilityTree,

		consensusStateManager: consensusStateManager,
		acceptanceDataStore:   acceptanceDataStore,
		blockIndex:            blockIndex,
		blockMessageStore:     blockMessageStore,
		blockStatusStore:      blockStatusStore,
	}
}

// BuildBlock builds a block over the current state, with the transactions
// selected by the given transactionSelector
func (bp *BlockProcessor) BuildBlock(scriptPublicKey []byte, extraData []byte,
	transactionSelector model.TransactionSelector) *appmessage.MsgBlock {

	return nil
}

// ValidateAndInsertBlock validates the given block and, if valid, applies it
// to the current state
func (bp *BlockProcessor) ValidateAndInsertBlock(block *appmessage.MsgBlock) error {
	return nil
}
