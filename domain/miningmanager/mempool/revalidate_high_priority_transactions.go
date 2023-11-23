package mempool

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool/model"
	"github.com/kaspanet/kaspad/infrastructure/logger"
)

func (mp *mempool) revalidateHighPriorityTransactions() ([]*externalapi.DomainTransaction, error) {
	type txNode struct {
		children          map[externalapi.DomainTransactionID]struct{}
		isInvalid         bool
		nonVisitedParents int
		tx                *model.MempoolTransaction
		visited           bool
	}

	onEnd := logger.LogAndMeasureExecutionTime(log, "revalidateHighPriorityTransactions")
	defer onEnd()

	// We revalidate transactions in topological order in case there are dependencies between them

	// Naturally transactions points to their dependencies, but since we want to start processing the dependencies
	// first, we build the opposite DAG. We initially fill `queue` with transactions with no dependencies.
	txDAG := make(map[externalapi.DomainTransactionID]*txNode)

	maybeAddNode := func(txID externalapi.DomainTransactionID) *txNode {
		if node, ok := txDAG[txID]; ok {
			return node
		}

		node := &txNode{
			children:          make(map[externalapi.DomainTransactionID]struct{}),
			isInvalid:         false,
			nonVisitedParents: 0,
			tx:                mp.transactionsPool.highPriorityTransactions[txID],
		}
		txDAG[txID] = node
		return node
	}

	queue := make([]*txNode, 0, len(mp.transactionsPool.highPriorityTransactions))
	for id, transaction := range mp.transactionsPool.highPriorityTransactions {
		node := &txNode{
			children:          make(map[externalapi.DomainTransactionID]struct{}),
			isInvalid:         false,
			nonVisitedParents: 0,
			tx:                transaction,
		}
		txDAG[id] = node

		parents := make(map[externalapi.DomainTransactionID]struct{})
		for _, input := range transaction.Transaction().Inputs {
			if _, ok := mp.transactionsPool.highPriorityTransactions[input.PreviousOutpoint.TransactionID]; !ok {
				continue
			}

			parents[input.PreviousOutpoint.TransactionID] = struct{}{} // To avoid duplicate parents, we first add it to a set and then count it
			maybeAddNode(input.PreviousOutpoint.TransactionID).children[id] = struct{}{}
		}
		node.nonVisitedParents = len(parents)

		if node.nonVisitedParents == 0 {
			queue = append(queue, node)
		}
	}

	validTransactions := []*externalapi.DomainTransaction{}

	// Now we iterate the DAG in topological order using BFS
	for len(queue) > 0 {
		var node *txNode
		node, queue = queue[0], queue[1:]

		if node.visited {
			continue
		}
		node.visited = true
		if node.isInvalid {
			continue
		}

		transaction := node.tx
		isValid, err := mp.revalidateTransaction(transaction)
		if err != nil {
			return nil, err
		}

		if !isValid {
			// Invalidate the offspring of this transaction
			invalidateQueue := []*txNode{node}
			for len(invalidateQueue) > 0 {
				var current *txNode
				current, invalidateQueue = invalidateQueue[0], invalidateQueue[1:]

				if current.isInvalid {
					continue
				}
				current.isInvalid = true
				for child := range current.children {
					invalidateQueue = append(invalidateQueue, txDAG[child])
				}
			}
			continue
		}

		for child := range node.children {
			childNode := txDAG[child]
			childNode.nonVisitedParents--
			if childNode.nonVisitedParents == 0 {
				queue = append(queue, txDAG[child])
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
