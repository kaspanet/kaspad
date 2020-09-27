package miningmanager

import "github.com/kaspanet/kaspad/app/appmessage"

type MiningManager interface {
	GetBlockTemplate() *appmessage.MsgBlock
	HandleNewBlock(block *appmessage.MsgBlock)
	ValidateAndInsertTransaction(transaction *appmessage.MsgTx) error
}

type miningManager struct {
}

func (mm *miningManager) GetBlockTemplate() *appmessage.MsgBlock {
	return nil
}

func (mm *miningManager) HandleNewBlock(block *appmessage.MsgBlock) {

}

func (mm *miningManager) ValidateAndInsertTransaction(transaction *appmessage.MsgTx) error {
	return nil
}
