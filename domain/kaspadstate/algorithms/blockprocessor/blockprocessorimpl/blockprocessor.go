package blockprocessorimpl

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/kaspadstate/algorithms/blockvalidator"
	"github.com/kaspanet/kaspad/domain/kaspadstate/algorithms/consensusstatemanager"
	"github.com/kaspanet/kaspad/domain/kaspadstate/algorithms/dagtopologymanager"
	"github.com/kaspanet/kaspad/domain/kaspadstate/algorithms/pruningmanager"
	"github.com/kaspanet/kaspad/domain/kaspadstate/algorithms/reachabilitytree"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures"
	"github.com/kaspanet/kaspad/domain/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
)

type BlockProcessor struct {
	dagParams       *dagconfig.Params
	databaseContext *dbaccess.DatabaseContext

	consensusStateManager consensusstatemanager.ConsensusStateManager
	pruningManager        pruningmanager.PruningManager
	blockValidator        blockvalidator.BlockValidator
	dagTopologyManager    dagtopologymanager.DAGTopologyManager
	reachabilityTree      reachabilitytree.ReachabilityTree
	acceptanceDataStore   datastructures.AcceptanceDataStore
	blockIndex            datastructures.BlockIndex
	blockMessageStore     datastructures.BlockMessageStore
	blockStatusStore      datastructures.BlockStatusStore
}

func New(
	dagParams *dagconfig.Params,
	databaseContext *dbaccess.DatabaseContext,
	consensusStateManager consensusstatemanager.ConsensusStateManager,
	pruningManager pruningmanager.PruningManager,
	blockValidator blockvalidator.BlockValidator,
	dagTopologyManager dagtopologymanager.DAGTopologyManager,
	reachabilityTree reachabilitytree.ReachabilityTree,
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

func (bp *BlockProcessor) BuildBlock(transactionSelector model.TransactionSelector) *appmessage.MsgBlock {
	return nil
}

func (bp *BlockProcessor) ValidateAndInsertBlock(block *appmessage.MsgBlock) error {
	return nil
}
