package mempool

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool/model"
)

func (mp *mempool) removeTransactions(transactions []*externalapi.DomainTransaction, removeRedeemers bool) error {
	for _, transaction := range transactions {
		err := mp.removeTransaction(consensushashing.TransactionID(transaction), removeRedeemers)
		if err != nil {
			return err
		}
	}
	return nil
}

func (mp *mempool) removeTransaction(transactionID *externalapi.DomainTransactionID, removeRedeemers bool) error {
	if _, ok := mp.orphansPool.allOrphans[*transactionID]; ok {
		return mp.orphansPool.removeOrphan(transactionID, true)
	}

	mempoolTransaction, ok := mp.transactionsPool.allTransactions[*transactionID]
	if !ok {
		return nil
	}

	transactionsToRemove := []*model.MempoolTransaction{mempoolTransaction}
	redeemers := mp.transactionsPool.getRedeemers(mempoolTransaction)
	if removeRedeemers {
		transactionsToRemove = append(transactionsToRemove, redeemers...)
	} else {
		for _, redeemer := range redeemers {
			redeemer.RemoveParentTransactionInPool(transactionID)
		}
	}

	for _, transactionToRemove := range transactionsToRemove {
		err := mp.removeTransactionFromSets(transactionToRemove, removeRedeemers)
		if err != nil {
			return err
		}
	}

	if removeRedeemers {
		err := mp.orphansPool.removeRedeemersOf(mempoolTransaction)
		if err != nil {
			return err
		}
	}

	return nil
}

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
