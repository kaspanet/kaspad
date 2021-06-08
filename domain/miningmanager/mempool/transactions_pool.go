package mempool

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool/model"
)

type transactionsPool struct {
	mempool                               *mempool
	allTransactions                       model.IDToTransaction
	highPriorityTransactions              model.IDToTransaction
	chainedTransactionsByPreviousOutpoint model.OutpointToTransaction
	transactionsByFeeRate                 model.TransactionsOrderedByFeeRate
	lastExpireScan                        uint64
}

func newTransactionsPool(mp *mempool) *transactionsPool {
	return &transactionsPool{
		mempool:                               mp,
		allTransactions:                       model.IDToTransaction{},
		highPriorityTransactions:              model.IDToTransaction{},
		chainedTransactionsByPreviousOutpoint: model.OutpointToTransaction{},
		transactionsByFeeRate:                 model.TransactionsOrderedByFeeRate{},
		lastExpireScan:                        0,
	}
}

func (tp *transactionsPool) addTransaction(transaction *externalapi.DomainTransaction, parentsInPool []*model.MempoolTransaction) error {
	panic("transactionsPool.addTransaction not implemented") // TODO (Mike)
}

func (tp *transactionsPool) addMempoolTransaction(transaction *model.MempoolTransaction) error {
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
		if mempoolTransaction.IsHighPriority {
			continue
		}

		// Remove all transactions whose addedAtDAAScore is older then transactionExpireIntervalDAAScore
		if virtualDAAScore-mempoolTransaction.AddedAtDAAScore > tp.mempool.config.transactionExpireIntervalDAAScore {
			err = tp.mempool.RemoveTransaction(mempoolTransaction.TransactionID())
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
		if len(mempoolTransaction.ParentsInPool) == 0 {
			result = append(result, mempoolTransaction.Transaction)
		}
	}

	return result
}
