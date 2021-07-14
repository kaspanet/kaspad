package flowcontext

import (
	"time"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/flows/transactionrelay"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
)

// TransactionIDPropagationInterval is the interval between transaction IDs propagations
const TransactionIDPropagationInterval = 500 * time.Millisecond

// AddTransaction adds transaction to the mempool and propagates it.
func (f *FlowContext) AddTransaction(tx *externalapi.DomainTransaction, allowOrphan bool) error {
	acceptedTransactions, err := f.Domain().MiningManager().ValidateAndInsertTransaction(tx, true, allowOrphan)
	if err != nil {
		return err
	}

	acceptedTransactionIDs := consensushashing.TransactionIDs(acceptedTransactions)
	return f.EnqueueTransactionIDsForPropagation(acceptedTransactionIDs)
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

// EnqueueTransactionIDsForPropagation add the given transactions IDs to a set of IDs to
// propagate. The IDs will be broadcast to all peers within a single transaction Inv message.
// The broadcast itself may happen only during a subsequent call to this method
func (f *FlowContext) EnqueueTransactionIDsForPropagation(transactionIDs []*externalapi.DomainTransactionID) error {
	f.transactionIDPropagationLock.Lock()
	defer f.transactionIDPropagationLock.Unlock()

	f.transactionIDsToPropagate = append(f.transactionIDsToPropagate, transactionIDs...)

	return f.maybePropagateTransactions()
}

func (f *FlowContext) maybePropagateTransactions() error {
	if time.Since(f.lastTransactionIDPropagationTime) < TransactionIDPropagationInterval &&
		len(f.transactionIDsToPropagate) < appmessage.MaxInvPerTxInvMsg {
		return nil
	}

	for len(f.transactionIDsToPropagate) > 0 {
		transactionIDsToBroadcast := f.transactionIDsToPropagate
		if len(transactionIDsToBroadcast) > appmessage.MaxInvPerTxInvMsg {
			transactionIDsToBroadcast = f.transactionIDsToPropagate[:len(transactionIDsToBroadcast)]
		}
		log.Infof("Transaction propagation: broadcasting %d transactions", len(transactionIDsToBroadcast))

		inv := appmessage.NewMsgInvTransaction(transactionIDsToBroadcast)
		err := f.Broadcast(inv)
		if err != nil {
			return err
		}

		f.transactionIDsToPropagate = f.transactionIDsToPropagate[len(transactionIDsToBroadcast):]
	}

	f.lastTransactionIDPropagationTime = time.Now()

	return nil
}
