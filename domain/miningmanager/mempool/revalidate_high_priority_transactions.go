package mempool

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool/model"
)

func (mp *mempool) revalidateHighPriorityTransactions() ([]*externalapi.DomainTransaction, error) {
	validTransactions := []*externalapi.DomainTransaction{}

	for _, transaction := range mp.transactionsPool.highPriorityTransactions {
		clearInputs(transaction)
		_, missingParents, err := mp.fillInputsAndGetMissingParents(transaction.Transaction())
		if err != nil {
			return nil, err
		}
		if len(missingParents) > 0 {
			err := mp.removeTransaction(transaction.TransactionID(), true)
			if err != nil {
				return nil, err
			}
			continue
		}
		validTransactions = append(validTransactions, transaction.Transaction())
	}

	return validTransactions, nil
}

func clearInputs(transaction *model.MempoolTransaction) {
	for _, input := range transaction.Transaction().Inputs {
		input.UTXOEntry = nil
	}
}
