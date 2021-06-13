package miningmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	miningmanagermodel "github.com/kaspanet/kaspad/domain/miningmanager/model"
)

// MiningManager creates block templates for mining as well as maintaining
// known transactions that have no yet been added to any block
type MiningManager interface {
	GetBlockTemplate(coinbaseData *externalapi.DomainCoinbaseData) (*externalapi.DomainBlock, error)
	GetTransaction(transactionID *externalapi.DomainTransactionID) (*externalapi.DomainTransaction, bool)
	AllTransactions() []*externalapi.DomainTransaction
	TransactionCount() int
	HandleNewBlockTransactions(txs []*externalapi.DomainTransaction) ([]*externalapi.DomainTransaction, error)
	ValidateAndInsertTransaction(transaction *externalapi.DomainTransaction, isHighPriority bool, allowOrphan bool) (
		acceptedTransactions []*externalapi.DomainTransaction, err error)
	RevalidateHighPriorityTransactions() (validTransactions []*externalapi.DomainTransaction, err error)
}

type miningManager struct {
	mempool              miningmanagermodel.Mempool
	blockTemplateBuilder miningmanagermodel.BlockTemplateBuilder
}

// GetBlockTemplate creates a block template for a miner to consume
func (mm *miningManager) GetBlockTemplate(coinbaseData *externalapi.DomainCoinbaseData) (*externalapi.DomainBlock, error) {
	return mm.blockTemplateBuilder.GetBlockTemplate(coinbaseData)
}

// HandleNewBlock handles the transactions for a new block that was just added to the DAG
func (mm *miningManager) HandleNewBlockTransactions(txs []*externalapi.DomainTransaction) ([]*externalapi.DomainTransaction, error) {
	return mm.mempool.HandleNewBlockTransactions(txs)
}

// ValidateAndInsertTransaction validates the given transaction, and
// adds it to the set of known transactions that have not yet been
// added to any block
func (mm *miningManager) ValidateAndInsertTransaction(transaction *externalapi.DomainTransaction,
	isHighPriority bool, allowOrphan bool) (acceptedTransactions []*externalapi.DomainTransaction, err error) {

	return mm.mempool.ValidateAndInsertTransaction(transaction, isHighPriority, allowOrphan)
}

func (mm *miningManager) GetTransaction(
	transactionID *externalapi.DomainTransactionID) (*externalapi.DomainTransaction, bool) {

	return mm.mempool.GetTransaction(transactionID)
}

func (mm *miningManager) AllTransactions() []*externalapi.DomainTransaction {
	return mm.mempool.AllTransactions()
}

func (mm *miningManager) TransactionCount() int {
	return mm.mempool.TransactionCount()
}

func (mm *miningManager) RevalidateHighPriorityTransactions() (
	validTransactions []*externalapi.DomainTransaction, err error) {

	return mm.mempool.RevalidateHighPriorityTransactions()
}
