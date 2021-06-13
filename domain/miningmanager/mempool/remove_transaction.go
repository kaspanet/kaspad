package mempool

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool/model"
)

func (mp *mempool) RemoveTransactions(transactions []*externalapi.DomainTransaction, removeRedeemers bool) error {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	return mp.removeTransactions(transactions, removeRedeemers)
}

func (mp *mempool) RemoveTransaction(transactionID *externalapi.DomainTransactionID, removeRedeemers bool) error {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	return mp.removeTransaction(transactionID, removeRedeemers)
}

// this function MUST be called with the mempool mutex locked for writes
func (mp *mempool) removeTransactions(transactions []*externalapi.DomainTransaction, removeRedeemers bool) error {
	for _, transaction := range transactions {
		err := mp.removeTransaction(consensushashing.TransactionID(transaction), removeRedeemers)
		if err != nil {
			return err
		}
	}
	return nil
}

// this function MUST be called with the mempool mutex locked for writes
func (mp *mempool) removeTransaction(transactionID *externalapi.DomainTransactionID, removeRedeemers bool) error {
	if _, ok := mp.orphansPool.allOrphans[*transactionID]; ok {
		return mp.orphansPool.removeOrphan(transactionID, true)
	}

	mempoolTransaction, ok := mp.transactionsPool.allTransactions[*transactionID]
	if !ok {
		return nil
	}

	transactionsToRemove := []*model.MempoolTransaction{mempoolTransaction}
	if removeRedeemers {
		redeemers := mp.transactionsPool.getRedeemers(mempoolTransaction)
		transactionsToRemove = append(transactionsToRemove, redeemers...)
	}

	for _, transactionToRemove := range transactionsToRemove {
		err := mp.removeTransactionFromSets(transactionToRemove, removeRedeemers)
		if err != nil {
			return err
		}
	}
	return nil
}

// this function MUST be called with the mempool mutex locked for writes
func (mp *mempool) removeTransactionFromSets(mempoolTransaction *model.MempoolTransaction, removeRedeemers bool) error {
	mp.mempoolUTXOSet.removeTransaction(mempoolTransaction)

	err := mp.transactionsPool.removeTransaction(mempoolTransaction)
	if err != nil {
		return err
	}

	err = mp.orphansPool.updateOrphansAfterTransactionRemoved(mempoolTransaction, removeRedeemers)
	if err != nil {
		return err
	}

	return nil
}
