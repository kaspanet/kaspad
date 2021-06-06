package mempool

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/dagconfig"
)

type mempool struct {
	config    *config
	consensus externalapi.Consensus

	mempoolUTXOSet   *mempoolUTXOSet
	transactionsPool *transactionsPool
	orphansPool      *orphansPool
}

func newMempool(consensus externalapi.Consensus, dagParams *dagconfig.Params) *mempool {
	mp := &mempool{
		config:    defaultConfig(dagParams),
		consensus: consensus,
	}

	mp.mempoolUTXOSet = newMempoolUTXOSet(mp)
	mp.transactionsPool = newTransactionsPool(mp)
	mp.orphansPool = newOrphansPool(mp)

	return mp
}

func (mp *mempool) ValidateAndInsertTransaction(transaction *externalapi.DomainTransaction, isHighPriority bool) (
	acceptedTransactions []*externalapi.DomainTransaction, err error) {

	panic("mempool.ValidateAndInsertTransaction not implemented") // TODO (Mike)
}

func (mp *mempool) HandleNewBlockTransactions(transactions []*externalapi.DomainTransaction) (
	acceptedOrphans []*externalapi.DomainTransaction, err error) {

	panic("mempool.HandleNewBlockTransactions not implemented") // TODO (Mike)
}

func (mp *mempool) RemoveTransaction(transactionID *externalapi.DomainTransactionID) error {
	panic("mempool.RemoveTransaction not implemented") // TODO (Mike)
}

func (mp *mempool) BlockCandidateTransactions() []*externalapi.DomainTransaction {
	return mp.transactionsPool.allReadyTransactions()
}

func (mp *mempool) RevalidateHighPriorityTransactions() (validTransactions []*externalapi.DomainTransaction, err error) {
	panic("mempool.RevalidateHighPriorityTransactions not implemented") // TODO (Mike)
}

func (mp *mempool) validateTransactionInIsolation(transaction *externalapi.DomainTransaction) error {
	panic("mempool.validateTransactionInIsolation not implemented") // TODO (Mike)
}

func (mp *mempool) validateTransactionInContext(transaction *externalapi.DomainTransaction) error {
	panic("mempool.validateTransactionInContext not implemented") // TODO (Mike)
}

func (mp *mempool) fillInputsAndGetMissingParents(transaction *externalapi.DomainTransaction) (
	parents []*mempoolTransaction, missingParents []externalapi.DomainTransactionID, err error) {

	panic("mempool.fillInputsAndGetMissingParents not implemented") // TODO (Mike)
}
