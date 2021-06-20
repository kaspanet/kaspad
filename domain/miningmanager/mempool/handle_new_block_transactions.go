package mempool

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
)

// this function MUST be called with the mempool mutex locked for writes
func (mp *mempool) handleNewBlockTransactions(blockTransactions []*externalapi.DomainTransaction) (
	[]*externalapi.DomainTransaction, error) {

	// Skip the coinbase transaction
	blockTransactions = blockTransactions[transactionhelper.CoinbaseTransactionIndex+1:]

	acceptedOrphans := []*externalapi.DomainTransaction{}
	for _, transaction := range blockTransactions {
		transactionID := consensushashing.TransactionID(transaction)
		err := mp.removeTransaction(transactionID, false)
		if err != nil {
			return nil, err
		}

		err = mp.removeDoubleSpends(transaction)
		if err != nil {
			return nil, err
		}

		err = mp.orphansPool.removeOrphan(transactionID, false)
		if err != nil {
			return nil, err
		}

		acceptedOrphansFromThisTransaction, err := mp.orphansPool.processOrphansAfterAcceptedTransaction(transaction)
		if err != nil {
			return nil, err
		}

		acceptedOrphans = append(acceptedOrphans, acceptedOrphansFromThisTransaction...)
	}
	err := mp.orphansPool.expireOrphanTransactions()
	if err != nil {
		return nil, err
	}
	err = mp.transactionsPool.expireOldTransactions()
	if err != nil {
		return nil, err
	}

	return acceptedOrphans, nil
}

// this function MUST be called with the mempool mutex locked for writes
func (mp *mempool) removeDoubleSpends(transaction *externalapi.DomainTransaction) error {
	for _, input := range transaction.Inputs {
		if redeemer, ok := mp.mempoolUTXOSet.transactionByPreviousOutpoint[input.PreviousOutpoint]; ok {
			err := mp.removeTransaction(redeemer.TransactionID(), true)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
