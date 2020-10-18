package miningmanager

import (
	consensusmodel "github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/miningmanager/model"
	"github.com/kaspanet/kaspad/util"
)

// MiningManager creates block templates for mining as well as maintaining
// known transactions that have no yet been added to any block
type MiningManager interface {
	GetBlockTemplate(payAddress util.Address, extraData []byte) *consensusmodel.DomainBlock
	HandleNewBlock(block *consensusmodel.DomainBlock)
	ValidateAndInsertTransaction(transaction *consensusmodel.DomainTransaction) error
}

type miningManager struct {
	mempool              model.Mempool
	blockTemplateBuilder model.BlockTemplateBuilder
}

// GetBlockTemplate creates a block template for a miner to consume
func (mm *miningManager) GetBlockTemplate(payAddress util.Address, extraData []byte) *consensusmodel.DomainBlock {
	return mm.blockTemplateBuilder.GetBlockTemplate(payAddress, extraData)
}

// HandleNewBlock handles a new block that was just added to the DAG
func (mm *miningManager) HandleNewBlock(block *consensusmodel.DomainBlock) {
	mm.mempool.HandleNewBlock(block)
}

// ValidateAndInsertTransaction validates the given transaction, and
// adds it to the set of known transactions that have not yet been
// added to any block
func (mm *miningManager) ValidateAndInsertTransaction(transaction *consensusmodel.DomainTransaction) error {
	return mm.mempool.ValidateAndInsertTransaction(transaction)
}
