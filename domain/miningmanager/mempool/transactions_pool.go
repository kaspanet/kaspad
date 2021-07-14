package mempool

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool/model"
	"time"
)

type transactionsPool struct {
	mempool                               *mempool
	allTransactions                       model.IDToTransactionMap
	highPriorityTransactions              model.IDToTransactionMap
	chainedTransactionsByPreviousOutpoint model.OutpointToTransactionMap
	transactionsOrderedByFeeRate          model.TransactionsOrderedByFeeRate
	lastExpireScanDAAScore                uint64
	lastExpireScanTime                    time.Time
}

func newTransactionsPool(mp *mempool) *transactionsPool {
	return &transactionsPool{
		mempool:                               mp,
		allTransactions:                       model.IDToTransactionMap{},
		highPriorityTransactions:              model.IDToTransactionMap{},
		chainedTransactionsByPreviousOutpoint: model.OutpointToTransactionMap{},
		transactionsOrderedByFeeRate:          model.TransactionsOrderedByFeeRate{},
		lastExpireScanDAAScore:                0,
		lastExpireScanTime:                    time.Now(),
	}
}

func (tp *transactionsPool) addTransaction(transaction *externalapi.DomainTransaction,
	parentTransactionsInPool model.OutpointToTransactionMap, isHighPriority bool) (*model.MempoolTransaction, error) {

	virtualDAAScore, err := tp.mempool.consensus.Consensus().GetVirtualDAAScore()
	if err != nil {
		return nil, err
	}

	mempoolTransaction := model.NewMempoolTransaction(
		transaction, parentTransactionsInPool, isHighPriority, virtualDAAScore)

	err = tp.addMempoolTransaction(mempoolTransaction)
	if err != nil {
		return nil, err
	}

	return mempoolTransaction, nil
}

func (tp *transactionsPool) addMempoolTransaction(transaction *model.MempoolTransaction) error {
	tp.allTransactions[*transaction.TransactionID()] = transaction

	for outpoint, parentTransactionInPool := range transaction.ParentTransactionsInPool() {
		tp.chainedTransactionsByPreviousOutpoint[outpoint] = parentTransactionInPool
	}

	tp.mempool.mempoolUTXOSet.addTransaction(transaction)

	err := tp.transactionsOrderedByFeeRate.Push(transaction)
	if err != nil {
		return err
	}

	if transaction.IsHighPriority() {
		tp.highPriorityTransactions[*transaction.TransactionID()] = transaction
	}

	return nil
}

func (tp *transactionsPool) removeTransaction(transaction *model.MempoolTransaction) error {
	delete(tp.allTransactions, *transaction.TransactionID())

	err := tp.transactionsOrderedByFeeRate.Remove(transaction)
	if err != nil {
		return err
	}

	delete(tp.highPriorityTransactions, *transaction.TransactionID())

	for outpoint := range transaction.ParentTransactionsInPool() {
		delete(tp.chainedTransactionsByPreviousOutpoint, outpoint)
	}

	return nil
}

func (tp *transactionsPool) expireOldTransactions() error {
	virtualDAAScore, err := tp.mempool.consensus.Consensus().GetVirtualDAAScore()
	if err != nil {
		return err
	}

	if virtualDAAScore-tp.lastExpireScanDAAScore < tp.mempool.config.TransactionExpireScanIntervalDAAScore ||
		time.Since(tp.lastExpireScanTime).Seconds() < float64(tp.mempool.config.TransactionExpireScanIntervalSeconds) {
		return nil
	}

	for _, mempoolTransaction := range tp.allTransactions {
		// Never expire high priority transactions
		if mempoolTransaction.IsHighPriority() {
			continue
		}

		// Remove all transactions whose addedAtDAAScore is older then TransactionExpireIntervalDAAScore
		daaScoreSinceAdded := virtualDAAScore - mempoolTransaction.AddedAtDAAScore()
		if daaScoreSinceAdded > tp.mempool.config.TransactionExpireIntervalDAAScore {
			log.Debugf("Removing transaction %s, because it expired. DAAScore moved by %d, expire interval: %d",
				mempoolTransaction.TransactionID(), daaScoreSinceAdded, tp.mempool.config.TransactionExpireIntervalDAAScore)
			err = tp.mempool.removeTransaction(mempoolTransaction.TransactionID(), true)
			if err != nil {
				return err
			}
		}
	}

	tp.lastExpireScanDAAScore = virtualDAAScore
	tp.lastExpireScanTime = time.Now()
	return nil
}

func (tp *transactionsPool) allReadyTransactions() []*externalapi.DomainTransaction {
	result := []*externalapi.DomainTransaction{}

	for _, mempoolTransaction := range tp.allTransactions {
		if len(mempoolTransaction.ParentTransactionsInPool()) == 0 {
			result = append(result, mempoolTransaction.Transaction())
		}
	}

	return result
}

func (tp *transactionsPool) getParentTransactionsInPool(
	transaction *externalapi.DomainTransaction) model.OutpointToTransactionMap {

	parentsTransactionsInPool := model.OutpointToTransactionMap{}

	for _, input := range transaction.Inputs {
		if transaction, ok := tp.allTransactions[input.PreviousOutpoint.TransactionID]; ok {
			parentsTransactionsInPool[input.PreviousOutpoint] = transaction
		}
	}

	return parentsTransactionsInPool
}

func (tp *transactionsPool) getRedeemers(transaction *model.MempoolTransaction) []*model.MempoolTransaction {
	queue := []*model.MempoolTransaction{transaction}
	redeemers := []*model.MempoolTransaction{}
	for len(queue) > 0 {
		var current *model.MempoolTransaction
		current, queue = queue[0], queue[1:]

		outpoint := externalapi.DomainOutpoint{TransactionID: *current.TransactionID()}
		for i := range current.Transaction().Outputs {
			outpoint.Index = uint32(i)
			if redeemerTransaction, ok := tp.chainedTransactionsByPreviousOutpoint[outpoint]; ok {
				queue = append(queue, redeemerTransaction)
				redeemers = append(redeemers, redeemerTransaction)
			}
		}
	}
	return redeemers
}

func (tp *transactionsPool) limitTransactionCount() error {
	currentIndex := 0

	for uint64(len(tp.allTransactions)) > tp.mempool.config.MaximumTransactionCount {
		var transactionToRemove *model.MempoolTransaction
		for {
			transactionToRemove = tp.transactionsOrderedByFeeRate.GetByIndex(currentIndex)
			if !transactionToRemove.IsHighPriority() {
				break
			}
			currentIndex++
			if currentIndex >= len(tp.allTransactions) {
				log.Warnf(
					"Number of high-priority transactions in mempool (%d) is higher than maximum allowed (%d)",
					len(tp.allTransactions), tp.mempool.config.MaximumTransactionCount)
				return nil
			}
		}

		log.Debugf("Removing transaction %s, because mempoolTransaction count (%d) exceeded the limit (%d)",
			transactionToRemove.TransactionID(), len(tp.allTransactions), tp.mempool.config.MaximumTransactionCount)
		err := tp.mempool.removeTransaction(transactionToRemove.TransactionID(), true)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tp *transactionsPool) getTransaction(transactionID *externalapi.DomainTransactionID) (*externalapi.DomainTransaction, bool) {
	if mempoolTransaction, ok := tp.allTransactions[*transactionID]; ok {
		return mempoolTransaction.Transaction(), true
	}
	return nil, false
}

func (tp *transactionsPool) getAllTransactions() []*externalapi.DomainTransaction {
	allTransactions := make([]*externalapi.DomainTransaction, 0, len(tp.allTransactions))
	for _, mempoolTransaction := range tp.allTransactions {
		allTransactions = append(allTransactions, mempoolTransaction.Transaction())
	}
	return allTransactions
}

func (tp *transactionsPool) transactionCount() int {
	return len(tp.allTransactions)
}
