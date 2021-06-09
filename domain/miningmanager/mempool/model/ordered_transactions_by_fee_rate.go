package model

import (
	"sort"

	"github.com/pkg/errors"
)

// TransactionsOrderedByFeeRate represents a set of MempoolTransactions ordered by their fee / mass rate
type TransactionsOrderedByFeeRate struct {
	slice []*MempoolTransaction
}

// Push inserts a transaction into the set, placing it in the correct place to preserve order
func (tobf *TransactionsOrderedByFeeRate) Push(transaction *MempoolTransaction) error {
	index, err := tobf.findTransactionIndex(transaction)
	if err != nil {
		return err
	}

	tobf.slice = append(tobf.slice[:index],
		append([]*MempoolTransaction{transaction}, tobf.slice[index:]...)...)

	return nil
}

// Remove removes the given transaction from the set.
// Returns an error if transaction does not exist in the set, or if the given transaction does not have mass
// and fee filled in.
func (tobf *TransactionsOrderedByFeeRate) Remove(transaction *MempoolTransaction) error {
	index, err := tobf.findTransactionIndex(transaction)
	if err != nil {
		return err
	}

	txID := transaction.TransactionID()
	if !tobf.slice[index].TransactionID().Equal(txID) {
		return errors.Errorf("Couldn't find %s in mp.orderedTransactionsByFeeRate", txID)
	}

	return tobf.RemoveAtIndex(index)
}

// RemoveAtIndex removes the transaction at the given index.
// Returns an error in case of out-of-bounds index.
func (tobf *TransactionsOrderedByFeeRate) RemoveAtIndex(index int) error {
	if index < 0 || index > len(tobf.slice)-1 {
		return errors.Errorf("Index %d is out of bound of this TransactionsOrderedByFeeRate", index)
	}
	tobf.slice = append(tobf.slice[:index], tobf.slice[index+1:]...)
	return nil
}

func (tobf *TransactionsOrderedByFeeRate) findTransactionIndex(transaction *MempoolTransaction) (int, error) {
	if transaction.Transaction().Fee == 0 || transaction.Transaction().Mass == 0 {
		return 0, errors.Errorf("findTxIndexInOrderedTransactionsByFeeRate expects a transaction with " +
			"populated fee and mass")
	}
	txID := transaction.TransactionID()
	txFeeRate := float64(transaction.Transaction().Fee) / float64(transaction.Transaction().Mass)

	return sort.Search(len(tobf.slice), func(i int) bool {
		iElement := tobf.slice[i]
		elementFeeRate := float64(iElement.Transaction().Fee) / float64(iElement.Transaction().Mass)
		if elementFeeRate > txFeeRate {
			return true
		}

		if elementFeeRate == txFeeRate && txID.LessOrEqual(iElement.TransactionID()) {
			return true
		}

		return false
	}), nil
}
