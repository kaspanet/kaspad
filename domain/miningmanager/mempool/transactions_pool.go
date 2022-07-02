package mempool

import (
	"time"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
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

func (tp *transactionsPool) allReadyTransactions(clone bool) []*externalapi.DomainTransaction {
	result := []*externalapi.DomainTransaction{}

	for _, mempoolTransaction := range tp.allTransactions {
		if len(mempoolTransaction.ParentTransactionsInPool()) == 0 {
			if clone {
				result = append(result, mempoolTransaction.Transaction().Clone())
			} else {
				result = append(result, mempoolTransaction.Transaction())
			}
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
			return mempoolTransaction.Transaction().Clone(), true
		}
		return mempoolTransaction.Transaction(), true
	}
	return nil, false
}

func (tp *transactionsPool) getTransactionsByAddresses(clone bool) (
	sending map[string]*externalapi.DomainTransaction,
	receiving map[string]*externalapi.DomainTransaction,
	err error) {
	sending = make(map[string]*externalapi.DomainTransaction)
	receiving = make(map[string]*externalapi.DomainTransaction)
	var transaction *externalapi.DomainTransaction
	for _, mempoolTransaction := range tp.allTransactions {
		if clone {
			transaction = mempoolTransaction.Transaction().Clone()
		} else {
			transaction = mempoolTransaction.Transaction()
		}
		for _, input := range transaction.Inputs {
			if input.UTXOEntry == nil { //this should be fixed
				return nil, nil, err
			}
			_, address, err := txscript.ExtractScriptPubKeyAddress(input.UTXOEntry.ScriptPublicKey(), tp.mempool.params)
			if err != nil {
				return nil, nil, err
			}
			if address == nil { //ignore none-standard script
				continue
			}
			sending[address.String()] = transaction
		}
		for _, output := range transaction.Outputs {
			_, address, err := txscript.ExtractScriptPubKeyAddress(output.ScriptPublicKey, tp.mempool.params)
			if err != nil {
				return nil, nil, err
			}
			if address == nil { //ignore none-standard script
				continue
			}
			receiving[address.String()] = transaction
		}
	}

	return sending, receiving, nil
}

func (tp *transactionsPool) getAllTransactions(clone bool) []*externalapi.DomainTransaction {
	allTransactions := make([]*externalapi.DomainTransaction, len(tp.allTransactions))
	i := 0
	for _, mempoolTransaction := range tp.allTransactions {
		if clone {
			allTransactions[i] = mempoolTransaction.Transaction().Clone()
		} else {
			allTransactions[i] = mempoolTransaction.Transaction()
		}
		i++
	}
	return allTransactions
}

func (tp *transactionsPool) transactionCount() int {
	return len(tp.allTransactions)
}
