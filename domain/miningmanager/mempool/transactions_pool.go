package mempool

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type outpointToTransaction map[externalapi.DomainOutpoint]*mempoolTransaction

type transactionsByFeeHeap []*mempoolTransaction

type transactionsPool struct {
	mempool                               *mempool
	allTransactions                       idToTransaction
	highPriorityTransactions              idToTransaction
	chainedTransactionsByPreviousOutpoint outpointToTransaction
	transactionsByFeeRate                 transactionsByFeeHeap
	lastExpireScan                        uint64
}

func newTransactionsPool(mp *mempool) *transactionsPool {
	return &transactionsPool{
		mempool:                               mp,
		allTransactions:                       idToTransaction{},
		highPriorityTransactions:              idToTransaction{},
		chainedTransactionsByPreviousOutpoint: outpointToTransaction{},
		transactionsByFeeRate:                 transactionsByFeeHeap{},
		lastExpireScan:                        0,
	}
}

func (tp *transactionsPool) addTransaction(transaction *externalapi.DomainTransaction, parentsInPool []*mempoolTransaction) error {
	panic("transactionsPool.addTransaction not implemented") // TODO (Mike)
}

func (tp *transactionsPool) addMempoolTransaction(transaction mempoolTransaction) error {
	panic("transactionsPool.addMempoolTransaction not implemented") // TODO (Mike)
}

func (tp *transactionsPool) expireOldTransactions() error {
	virtualDAAScore, err := tp.mempool.virtualDAAScore()
	if err != nil {
		return err
	}

	if virtualDAAScore-tp.lastExpireScan < tp.mempool.config.transactionExpireScanIntervalDAAScore {
		return nil
	}

	for _, mempoolTransaction := range tp.allTransactions {
		// Never expire high priority transactions
		if mempoolTransaction.isHighPriority {
			continue
		}

		// Remove all transactions whose addedAtDAAScore is older then transactionExpireIntervalDAAScore
		if virtualDAAScore-mempoolTransaction.addAtDAAScore > tp.mempool.config.transactionExpireIntervalDAAScore {
			err = tp.mempool.RemoveTransaction(mempoolTransaction.transactionID())
			if err != nil {
				return err
			}
		}
	}

	tp.lastExpireScan = virtualDAAScore
	return nil
}

func (tp *transactionsPool) allReadyTransactions() []*externalapi.DomainTransaction {
	result := []*externalapi.DomainTransaction{}

	for _, mempoolTransaction := range tp.allTransactions {
		if len(mempoolTransaction.parentsInPool) == 0 {
			result = append(result, mempoolTransaction.transaction)
		}
	}

	return result
}
