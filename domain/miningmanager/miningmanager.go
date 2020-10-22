package miningmanager

import (
	consensusexternalapi "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	miningmanagermodel "github.com/kaspanet/kaspad/domain/miningmanager/model"
)

// MiningManager creates block templates for mining as well as maintaining
// known transactions that have no yet been added to any block
type MiningManager interface {
	GetBlockTemplate(coinbaseData *consensusexternalapi.DomainCoinbaseData) *consensusexternalapi.DomainBlock
	HandleNewBlock(block *consensusexternalapi.DomainBlock)
	ValidateAndInsertTransaction(transaction *consensusexternalapi.DomainTransaction) error
}

type miningManager struct {
	mempool              miningmanagermodel.Mempool
	blockTemplateBuilder miningmanagermodel.BlockTemplateBuilder
}

// GetBlockTemplate creates a block template for a miner to consume
func (mm *miningManager) GetBlockTemplate(coinbaseData *consensusexternalapi.DomainCoinbaseData) *consensusexternalapi.DomainBlock {
	return mm.blockTemplateBuilder.GetBlockTemplate(coinbaseData)
}

// HandleNewBlock handles a new block that was just added to the DAG
func (mm *miningManager) HandleNewBlock(block *consensusexternalapi.DomainBlock) {
	mm.mempool.HandleNewBlock(block)
}

// ValidateAndInsertTransaction validates the given transaction, and
// adds it to the set of known transactions that have not yet been
// added to any block
func (mm *miningManager) ValidateAndInsertTransaction(transaction *consensusexternalapi.DomainTransaction) error {
	return mm.mempool.ValidateAndInsertTransaction(transaction)
}
