package mempool

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

type outpointToTransaction map[externalapi.DomainOutpoint]*mempoolTransaction

type transactionsByFeeHeap []*mempoolTransaction

type transactionsPool struct {
	mempool                               *mempool
	allTransactions                       idToTransaction
	highPriorityTransactions              idToTransaction
	chainedTransactionsByPreviousOutpoint outpointToTransaction
	transactionsByFeeRate                 transactionsByFeeHeap
}

func newTransactionsPool(mp *mempool) *transactionsPool {
	return &transactionsPool{
		mempool:                               mp,
		allTransactions:                       idToTransaction{},
		highPriorityTransactions:              idToTransaction{},
		chainedTransactionsByPreviousOutpoint: outpointToTransaction{},
		transactionsByFeeRate:                 transactionsByFeeHeap{},
	}
}

func (tp *transactionsPool) addTransaction(transaction *externalapi.DomainTransaction, parentsInPool []*mempoolTransaction) error {
	panic("transactionsPool.addTransaction not implemented") // TODO (Mike)
}

func (tp *transactionsPool) addMempoolTransaction(transaction mempoolTransaction) error {
	panic("transactionsPool.addMempoolTransaction not implemented") // TODO (Mike)
}

func (tp *transactionsPool) expireOldTransactions() error {
	panic("transactionsPool.expireOldTransactions not implemented") // TODO (Mike)
}
