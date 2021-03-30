package flowcontext

import (
	"time"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/flows/transactionrelay"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
)

// AddTransaction adds transaction to the mempool and propagates it.
func (f *FlowContext) AddTransaction(tx *externalapi.DomainTransaction) error {
	f.transactionsToRebroadcastLock.Lock()
	defer f.transactionsToRebroadcastLock.Unlock()

	err := f.Domain().MiningManager().ValidateAndInsertTransaction(tx, false)
	if err != nil {
		return err
	}

	transactionID := consensushashing.TransactionID(tx)
	f.transactionsToRebroadcast[*transactionID] = tx
	inv := appmessage.NewMsgInvTransaction([]*externalapi.DomainTransactionID{transactionID})
	return f.Broadcast(inv)
}

func (f *FlowContext) updateTransactionsToRebroadcast(addedBlocks []*externalapi.DomainBlock) {
	f.transactionsToRebroadcastLock.Lock()
	defer f.transactionsToRebroadcastLock.Unlock()

	for _, block := range addedBlocks {
		// Note: if a transaction is included in the DAG but not accepted,
		// it won't be rebroadcast anymore, although it is not included in
		// the UTXO set
		for _, tx := range block.Transactions {
			delete(f.transactionsToRebroadcast, *consensushashing.TransactionID(tx))
		}
	}
}

func (f *FlowContext) shouldRebroadcastTransactions() bool {
	const rebroadcastInterval = 30 * time.Second
	return time.Since(f.lastRebroadcastTime) > rebroadcastInterval
}

func (f *FlowContext) txIDsToRebroadcast() []*externalapi.DomainTransactionID {
	f.transactionsToRebroadcastLock.Lock()
	defer f.transactionsToRebroadcastLock.Unlock()

	txIDs := make([]*externalapi.DomainTransactionID, len(f.transactionsToRebroadcast))
	i := 0
	for _, tx := range f.transactionsToRebroadcast {
		txIDs[i] = consensushashing.TransactionID(tx)
		i++
	}
	return txIDs
}

// SharedRequestedTransactions returns a *transactionrelay.SharedRequestedTransactions for sharing
// data about requested transactions between different peers.
func (f *FlowContext) SharedRequestedTransactions() *transactionrelay.SharedRequestedTransactions {
	return f.sharedRequestedTransactions
}

// OnTransactionAddedToMempool notifies the handler function that a transaction
// has been added to the mempool
func (f *FlowContext) OnTransactionAddedToMempool() {
	if f.onTransactionAddedToMempoolHandler != nil {
		f.onTransactionAddedToMempoolHandler()
	}
}
