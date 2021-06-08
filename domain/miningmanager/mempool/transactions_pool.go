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

func (tp *transactionsPool) addTransaction(transaction *externalapi.DomainTransaction,
	parentTransactionsInPool model.OutpointToTransaction, isHighPriority bool) error {

	virtualDAAScore, err := tp.mempool.virtualDAAScore()
	if err != nil {
		return err
	}

	mempoolTransaction := &model.MempoolTransaction{
		Transaction:              transaction,
		ParentTransactionsInPool: parentTransactionsInPool,
		IsHighPriority:           isHighPriority,
		AddedAtDAAScore:          virtualDAAScore,
	}

	return tp.addMempoolTransaction(mempoolTransaction)
}

func (tp *transactionsPool) addMempoolTransaction(transaction *model.MempoolTransaction) error {
	tp.allTransactions[*transaction.TransactionID()] = transaction

	for outpoint, parentTransactionInPool := range transaction.ParentTransactionsInPool {
		tp.chainedTransactionsByPreviousOutpoint[outpoint] = parentTransactionInPool
	}

	tp.mempool.mempoolUTXOSet.addTransaction(transaction)

	err := tp.transactionsByFeeRate.Push(transaction)
	if err != nil {
		return err
	}

	if transaction.IsHighPriority {
		tp.highPriorityTransactions[*transaction.TransactionID()] = transaction
	}

	return nil
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
		if len(mempoolTransaction.ParentTransactionsInPool) == 0 {
			result = append(result, mempoolTransaction.Transaction)
		}
	}

	return result
}

func (tp *transactionsPool) getParentTransactionsInPool(
	transaction *externalapi.DomainTransaction) model.OutpointToTransaction {

	parentsTransactionsInPool := model.OutpointToTransaction{}

	for _, input := range transaction.Inputs {
		transaction := tp.allTransactions[input.PreviousOutpoint.TransactionID]
		parentsTransactionsInPool[input.PreviousOutpoint] = transaction
	}

	return parentsTransactionsInPool
}
