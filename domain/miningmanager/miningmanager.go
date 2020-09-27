package miningmanager

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/miningmanager/blocktemplatebuilder"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool"
)

type MiningManager interface {
	GetBlockTemplate() *appmessage.MsgBlock
	HandleNewBlock(block *appmessage.MsgBlock)
	ValidateAndInsertTransaction(transaction *appmessage.MsgTx) error
}

type miningManager struct {
	mempool              mempool.Mempool
	blockTemplateBuilder blocktemplatebuilder.BlockTemplateBuilder
}

func (mm *miningManager) GetBlockTemplate() *appmessage.MsgBlock {
	return mm.blockTemplateBuilder.GetBlockTemplate()
}

func (mm *miningManager) HandleNewBlock(block *appmessage.MsgBlock) {
	mm.mempool.HandleNewBlock(block)
}

func (mm *miningManager) ValidateAndInsertTransaction(transaction *appmessage.MsgTx) error {
	return mm.mempool.ValidateAndInsertTransaction(transaction)
}
