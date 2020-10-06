package blockprocessor

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/processes"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
)

// BlockProcessor is responsible for processing incoming blocks
// and creating blocks from the current state
type BlockProcessor struct {
	dagParams       *dagconfig.Params
	databaseContext *dbaccess.DatabaseContext

	consensusStateManager processes.ConsensusStateManager
	pruningManager        processes.PruningManager
	blockValidator        processes.BlockValidator
	dagTopologyManager    processes.DAGTopologyManager
	reachabilityTree      processes.ReachabilityTree
	acceptanceDataStore   datastructures.AcceptanceDataStore
	blockIndex            datastructures.BlockIndex
	blockMessageStore     datastructures.BlockMessageStore
	blockStatusStore      datastructures.BlockStatusStore
}

// New instantiates a new BlockProcessor
func New(
	dagParams *dagconfig.Params,
	databaseContext *dbaccess.DatabaseContext,
	consensusStateManager processes.ConsensusStateManager,
	pruningManager processes.PruningManager,
	blockValidator processes.BlockValidator,
	dagTopologyManager processes.DAGTopologyManager,
	reachabilityTree processes.ReachabilityTree,
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
func (bp *BlockProcessor) BuildBlock(coinbaseScriptPublicKey []byte, coinbaseExtraData []byte,
	transactionSelector model.TransactionSelector) *appmessage.MsgBlock {

	return nil
}

// ValidateAndInsertBlock validates the given block and, if valid, applies it
// to the current state
func (bp *BlockProcessor) ValidateAndInsertBlock(block *appmessage.MsgBlock) error {
	return nil
}

// SetOnBlockAddedToDAGHandler set the onBlockAddedToDAGHandler for the block processor
func (bp *BlockProcessor) SetOnBlockAddedToDAGHandler(onBlockAddedToDAGHandler model.OnBlockAddedToDAGHandler) {

}

// SetOnChainChangedHandler set the onBlockAddedToDAGHandler for the block processor
func (bp *BlockProcessor) SetOnChainChangedHandler(onChainChangedHandler model.OnChainChangedHandler) {

}

// SetOnFinalityConflictHandler set the onBlockAddedToDAGHandler for the block processor
func (bp *BlockProcessor) SetOnFinalityConflictHandler(onFinalityConflictHandler model.OnFinalityConflictHandler) {

}
