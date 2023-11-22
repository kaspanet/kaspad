package mempool

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool/model"
	"github.com/kaspanet/kaspad/infrastructure/logger"
)

func (mp *mempool) revalidateHighPriorityTransactions() ([]*externalapi.DomainTransaction, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "revalidateHighPriorityTransactions")
	defer onEnd()

	// We revalidate transactions in topological order in case there are dependencies between them

	// Naturally transactions points to their dependencies, but since we want to start processing the dependencies
	// first, we build the opposite DAG. We initially fill `queue` with transactions with no dependencies.
	childrenByID := make(map[externalapi.DomainTransactionID]map[externalapi.DomainTransactionID]struct{})
	queue := make([]externalapi.DomainTransactionID, 0, len(mp.transactionsPool.highPriorityTransactions))
	for id, transaction := range mp.transactionsPool.highPriorityTransactions {
		hasParents := false
		for _, input := range transaction.Transaction().Inputs {
			if _, ok := mp.transactionsPool.highPriorityTransactions[input.PreviousOutpoint.TransactionID]; !ok {
				continue
			}

			hasParents = true
			if _, ok := childrenByID[input.PreviousOutpoint.TransactionID]; !ok {
				childrenByID[input.PreviousOutpoint.TransactionID] = make(map[externalapi.DomainTransactionID]struct{})
			}

			childrenByID[input.PreviousOutpoint.TransactionID][id] = struct{}{}
		}

		if !hasParents {
			queue = append(queue, id)
		}
	}

	invalidTransactions := make(map[externalapi.DomainTransactionID]struct{})
	visited := make(map[externalapi.DomainTransactionID]struct{})
	validTransactions := []*externalapi.DomainTransaction{}

	// Now we iterate the DAG in topological order using BFS
	for len(queue) > 0 {
		var txID externalapi.DomainTransactionID
		txID, queue = queue[0], queue[1:]

		if _, ok := visited[txID]; ok {
			continue
		}
		visited[txID] = struct{}{}

		if _, ok := invalidTransactions[txID]; ok {
			continue
		}

		transaction := mp.transactionsPool.highPriorityTransactions[txID]
		isValid, err := mp.revalidateTransaction(transaction)
		if err != nil {
			return nil, err
		}

		if !isValid {
			// Invalidate the offspring of this transaction
			invalidateQueue := []externalapi.DomainTransactionID{txID}
			for len(invalidateQueue) > 0 {
				var current externalapi.DomainTransactionID
				current, invalidateQueue = invalidateQueue[0], invalidateQueue[1:]

				if _, ok := invalidTransactions[current]; ok {
					continue
				}

				invalidTransactions[current] = struct{}{}
				if children, ok := childrenByID[current]; ok {
					for child := range children {
						invalidateQueue = append(invalidateQueue, child)
					}
				}
			}
			continue
		}

		if children, ok := childrenByID[txID]; ok {
			for child := range children {
				queue = append(queue, child)
			}
		}

		validTransactions = append(validTransactions, transaction.Transaction().Clone())
	}

	return validTransactions, nil
}

func (mp *mempool) revalidateTransaction(transaction *model.MempoolTransaction) (isValid bool, err error) {
	clearInputs(transaction)

	_, missingParents, err := mp.fillInputsAndGetMissingParents(transaction.Transaction())
	if err != nil {
		return false, err
	}
	if len(missingParents) > 0 {
		log.Debugf("Removing transaction %s, it failed revalidation", transaction.TransactionID())
		err := mp.removeTransaction(transaction.TransactionID(), true)
		if err != nil {
			return false, err
		}
		return false, nil
	}

	return true, nil
}

func clearInputs(transaction *model.MempoolTransaction) {
	for _, input := range transaction.Transaction().Inputs {
		input.UTXOEntry = nil
	}
}
