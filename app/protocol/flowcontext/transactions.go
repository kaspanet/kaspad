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
	_, err := f.Domain().MiningManager().ValidateAndInsertTransaction(tx, true, false)
	if err != nil {
		return err
	}

	transactionID := consensushashing.TransactionID(tx)
	inv := appmessage.NewMsgInvTransaction([]*externalapi.DomainTransactionID{transactionID})

	return f.Broadcast(inv)
}

func (f *FlowContext) shouldRebroadcastTransactions() bool {
	const rebroadcastInterval = 30 * time.Second
	return time.Since(f.lastRebroadcastTime) > rebroadcastInterval
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
