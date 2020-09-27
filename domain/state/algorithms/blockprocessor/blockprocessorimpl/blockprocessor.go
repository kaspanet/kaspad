package blockprocessorimpl

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/model"
	"github.com/kaspanet/kaspad/domain/state/algorithms/blockvalidator"
	"github.com/kaspanet/kaspad/domain/state/algorithms/consensusstatemanager"
	"github.com/kaspanet/kaspad/domain/state/algorithms/dagtopologymanager"
	"github.com/kaspanet/kaspad/domain/state/algorithms/pruningmanager"
	"github.com/kaspanet/kaspad/domain/state/algorithms/reachabilitytree"
	"github.com/kaspanet/kaspad/domain/state/datastructures/acceptancedatastore"
	"github.com/kaspanet/kaspad/domain/state/datastructures/blockindex"
	"github.com/kaspanet/kaspad/domain/state/datastructures/blockmessagestore"
	"github.com/kaspanet/kaspad/domain/state/datastructures/blockstatusstore"
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
	acceptanceDataStore   acceptancedatastore.AcceptanceDataStore
	blockIndex            blockindex.BlockIndex
	blockMessageStore     blockmessagestore.BlockMessageStore
	blockStatusStore      blockstatusstore.BlockStatusStore
}

func New(
	dagParams *dagconfig.Params,
	databaseContext *dbaccess.DatabaseContext,
	consensusStateManager consensusstatemanager.ConsensusStateManager,
	pruningManager pruningmanager.PruningManager,
	blockValidator blockvalidator.BlockValidator,
	dagTopologyManager dagtopologymanager.DAGTopologyManager,
	reachabilityTree reachabilitytree.ReachabilityTree,
	acceptanceDataStore acceptancedatastore.AcceptanceDataStore,
	blockIndex blockindex.BlockIndex,
	blockMessageStore blockmessagestore.BlockMessageStore,
	blockStatusStore blockstatusstore.BlockStatusStore) *BlockProcessor {

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
