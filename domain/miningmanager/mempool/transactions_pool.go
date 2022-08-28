package mempool

import (
	"time"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool/model"
)

type transactionsPool struct {
	mempool                       *mempool
	allTransactions               model.IDToTransactionMap
	highPriorityTransactions      model.IDToTransactionMap
	chainedTransactionsByParentID model.IDToTransactionsSliceMap
	transactionsOrderedByFeeRate  model.TransactionsOrderedByFeeRate
	lastExpireScanDAAScore        uint64
	lastExpireScanTime            time.Time
}

func newTransactionsPool(mp *mempool) *transactionsPool {
	return &transactionsPool{
		mempool:                       mp,
		allTransactions:               model.IDToTransactionMap{},
		highPriorityTransactions:      model.IDToTransactionMap{},
		chainedTransactionsByParentID: model.IDToTransactionsSliceMap{},
		transactionsOrderedByFeeRate:  model.TransactionsOrderedByFeeRate{},
		lastExpireScanDAAScore:        0,
		lastExpireScanTime:            time.Now(),
	}
}

func (tp *transactionsPool) addTransaction(transaction *externalapi.DomainTransaction,
	parentTransactionsInPool model.IDToTransactionMap, isHighPriority bool) (*model.MempoolTransaction, error) {

	virtualDAAScore, err := tp.mempool.consensusReference.Consensus().GetVirtualDAAScore()
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

	for _, parentTransactionInPool := range transaction.ParentTransactionsInPool() {
		parentTransactionID := *parentTransactionInPool.TransactionID()
		if tp.chainedTransactionsByParentID[parentTransactionID] == nil {
			tp.chainedTransactionsByParentID[parentTransactionID] = []*model.MempoolTransaction{}
		}
		tp.chainedTransactionsByParentID[parentTransactionID] =
			append(tp.chainedTransactionsByParentID[parentTransactionID], transaction)
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
		if errors.Is(err, model.ErrTransactionNotFound) {
			log.Errorf("Transaction %s not found in tp.transactionsOrderedByFeeRate. This should never happen but sometime does",
				transaction.TransactionID())
		} else {
			return err
		}
	}

	delete(tp.highPriorityTransactions, *transaction.TransactionID())

	delete(tp.chainedTransactionsByParentID, *transaction.TransactionID())

	return nil
}

func (tp *transactionsPool) expireOldTransactions() error {
	virtualDAAScore, err := tp.mempool.consensusReference.Consensus().GetVirtualDAAScore()
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
			result = append(result, mempoolTransaction.Transaction().Clone()) //this pointer leaves the mempool, and gets its utxo set to nil, hence we clone.
		}
	}

	return result
}

func (tp *transactionsPool) getParentTransactionsInPool(
	transaction *externalapi.DomainTransaction) model.IDToTransactionMap {

	parentsTransactionsInPool := model.IDToTransactionMap{}

	for _, input := range transaction.Inputs {
		if transaction, ok := tp.allTransactions[input.PreviousOutpoint.TransactionID]; ok {
			parentsTransactionsInPool[*transaction.TransactionID()] = transaction
		}
	}

	return parentsTransactionsInPool
}

func (tp *transactionsPool) getRedeemers(transaction *model.MempoolTransaction) []*model.MempoolTransaction {
	stack := []*model.MempoolTransaction{transaction}
	redeemers := []*model.MempoolTransaction{}
	for len(stack) > 0 {
		var current *model.MempoolTransaction
		last := len(stack) - 1
		current, stack = stack[last], stack[:last]

		for _, redeemerTransaction := range tp.chainedTransactionsByParentID[*current.TransactionID()] {
			stack = append(stack, redeemerTransaction)
			redeemers = append(redeemers, redeemerTransaction)
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
		if currentIndex >= len(tp.allTransactions) {
			break
		}
	}
	return nil
}

func (tp *transactionsPool) getTransaction(transactionID *externalapi.DomainTransactionID, clone bool) (*externalapi.DomainTransaction, bool) {
	if mempoolTransaction, ok := tp.allTransactions[*transactionID]; ok {
		if clone {
			return mempoolTransaction.Transaction().Clone(), true //this pointer leaves the mempool, hence we clone.
		}
		return mempoolTransaction.Transaction(), true
	}
	return nil, false
}

func (tp *transactionsPool) getTransactionsByAddresses() (
	sending model.ScriptPublicKeyStringToDomainTransaction,
	receiving model.ScriptPublicKeyStringToDomainTransaction,
	err error) {
	sending = make(model.ScriptPublicKeyStringToDomainTransaction, tp.transactionCount())
	receiving = make(model.ScriptPublicKeyStringToDomainTransaction, tp.transactionCount())
	var transaction *externalapi.DomainTransaction
	for _, mempoolTransaction := range tp.allTransactions {
		transaction = mempoolTransaction.Transaction().Clone() //this pointer leaves the mempool, hence we clone.
		for _, input := range transaction.Inputs {
			if input.UTXOEntry == nil {
				return nil, nil, errors.Errorf("Mempool transaction %s is missing an UTXOEntry. This should be fixed, and not happen", consensushashing.TransactionID(transaction).String())
			}
			sending[input.UTXOEntry.ScriptPublicKey().String()] = transaction
		}
		for _, output := range transaction.Outputs {
			receiving[output.ScriptPublicKey.String()] = transaction
		}
	}
	return sending, receiving, nil
}

func (tp *transactionsPool) getAllTransactions() []*externalapi.DomainTransaction {
	allTransactions := make([]*externalapi.DomainTransaction, len(tp.allTransactions))
	i := 0
	for _, mempoolTransaction := range tp.allTransactions {
		allTransactions[i] = mempoolTransaction.Transaction().Clone() //this pointer leaves the mempool, hence we clone.
		i++
	}
	return allTransactions
}

func (tp *transactionsPool) transactionCount() int {
	return len(tp.allTransactions)
}
