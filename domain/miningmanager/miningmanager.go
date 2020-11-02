package miningmanager

import (
	consensusexternalapi "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	miningmanagermodel "github.com/kaspanet/kaspad/domain/miningmanager/model"
)

// MiningManager creates block templates for mining as well as maintaining
// known transactions that have no yet been added to any block
type MiningManager interface {
	GetBlockTemplate(coinbaseData *consensusexternalapi.DomainCoinbaseData) *consensusexternalapi.DomainBlock
	HandleNewBlockTransactions(txs []*consensusexternalapi.DomainTransaction)
	ValidateAndInsertTransaction(transaction *consensusexternalapi.DomainTransaction, allowOrphan bool) error
}

type miningManager struct {
	mempool              miningmanagermodel.Mempool
	blockTemplateBuilder miningmanagermodel.BlockTemplateBuilder
}

// GetBlockTemplate creates a block template for a miner to consume
func (mm *miningManager) GetBlockTemplate(coinbaseData *consensusexternalapi.DomainCoinbaseData) *consensusexternalapi.DomainBlock {
	return mm.blockTemplateBuilder.GetBlockTemplate(coinbaseData)
}

// HandleNewBlock handles the transactions for a new block that was just added to the DAG
func (mm *miningManager) HandleNewBlockTransactions(txs []*consensusexternalapi.DomainTransaction) {
	mm.mempool.HandleNewBlockTransactions(txs)
}

// ValidateAndInsertTransaction validates the given transaction, and
// adds it to the set of known transactions that have not yet been
// added to any block
func (mm *miningManager) ValidateAndInsertTransaction(transaction *consensusexternalapi.DomainTransaction, allowOrphan bool) error {
	return mm.mempool.ValidateAndInsertTransaction(transaction, allowOrphan)
}
