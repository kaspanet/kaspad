package miningmanager

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/util"
)

// MiningManager creates block templates for mining as well as maintaining
// known transactions that have no yet been added to any block
type MiningManager interface {
	GetBlockTemplate(payAddress util.Address, extraData []byte) *appmessage.MsgBlock
	HandleNewBlock(block *appmessage.MsgBlock)
	ValidateAndInsertTransaction(transaction *appmessage.MsgTx) error
}

type miningManager struct {
	mempool              Mempool
	blockTemplateBuilder BlockTemplateBuilder
}

// GetBlockTemplate creates a block template for a miner to consume
func (mm *miningManager) GetBlockTemplate(payAddress util.Address, extraData []byte) *appmessage.MsgBlock {
	return mm.blockTemplateBuilder.GetBlockTemplate(payAddress, extraData)
}

// HandleNewBlock handles a new block that was just added to the DAG
func (mm *miningManager) HandleNewBlock(block *appmessage.MsgBlock) {
	mm.mempool.HandleNewBlock(block)
}

// ValidateAndInsertTransaction validates the given transaction, and
// adds it to the set of known transactions that have not yet been
// added to any block
func (mm *miningManager) ValidateAndInsertTransaction(transaction *appmessage.MsgTx) error {
	return mm.mempool.ValidateAndInsertTransaction(transaction)
}
