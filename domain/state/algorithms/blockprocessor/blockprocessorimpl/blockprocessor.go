package blockprocessorimpl

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/model"
	"github.com/kaspanet/kaspad/domain/state/algorithms/consensusstatemanager"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
)

type BlockProcessor struct {
	dagParams       *dagconfig.Params
	databaseContext *dbaccess.DatabaseContext

	consensusStateManager consensusstatemanager.ConsensusStateManager
}

func New(dagParams *dagconfig.Params, databaseContext *dbaccess.DatabaseContext, consensusStateManager consensusstatemanager.ConsensusStateManager) *BlockProcessor {
	return &BlockProcessor{
		dagParams:       dagParams,
		databaseContext: databaseContext,

		consensusStateManager: consensusStateManager,
	}
}

func (bp *BlockProcessor) BuildBlock(transactionSelector model.TransactionSelector) *appmessage.MsgBlock {
	return nil
}

func (bp *BlockProcessor) ValidateAndInsertBlock(block *appmessage.MsgBlock) error {
	return nil
}
