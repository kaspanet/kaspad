package mempool

import (
	"sort"

	"github.com/pkg/errors"
)

type transactionsOrderedByFee struct {
	slice []*mempoolTransaction
}

func (tobf *transactionsOrderedByFee) findTransactionIndex(transaction *mempoolTransaction) (int, error) {
	if transaction.transaction.Fee == 0 || transaction.transaction.Mass == 0 {
		return 0, errors.Errorf("findTxIndexInOrderedTransactionsByFeeRate expects a transaction with " +
			"populated fee and mass")
	}
	txID := transaction.transactionID()
	txFeeRate := float64(transaction.transaction.Fee) / float64(transaction.transaction.Mass)

	return sort.Search(len(tobf.slice), func(i int) bool {
		iElement := tobf.slice[i]
		elementFeeRate := float64(iElement.transaction.Fee) / float64(iElement.transaction.Mass)
		if elementFeeRate > txFeeRate {
			return true
		}

		if elementFeeRate == txFeeRate && txID.LessOrEqual(iElement.transactionID()) {
			return true
		}

		return false
	}), nil
}

func (tobf *transactionsOrderedByFee) push(transaction *mempoolTransaction) error {
	index, err := tobf.findTransactionIndex(transaction)
	if err != nil {
		return err
	}

	tobf.slice = append(tobf.slice[:index],
		append([]*mempoolTransaction{transaction}, tobf.slice[index:]...)...)

	return nil
}

func (tobf *transactionsOrderedByFee) remove(transaction *mempoolTransaction) error {
	index, err := tobf.findTransactionIndex(transaction)
	if err != nil {
		return err
	}

	txID := transaction.transactionID()
	if !tobf.slice[index].transactionID().Equal(txID) {
		return errors.Errorf("Couldn't find %s in mp.orderedTransactionsByFeeRate", txID)
	}

	tobf.removeAtIndex(index)
	return nil
}

func (tobf *transactionsOrderedByFee) removeAtIndex(index int) {
	tobf.slice = append(tobf.slice[:index], tobf.slice[index+1:]...)
}
