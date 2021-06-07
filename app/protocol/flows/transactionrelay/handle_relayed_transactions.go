package transactionrelay

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/common"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool_old"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
)

// TransactionsRelayContext is the interface for the context needed for the
// HandleRelayedTransactions and HandleRequestedTransactions flows.
type TransactionsRelayContext interface {
	NetAdapter() *netadapter.NetAdapter
	Domain() domain.Domain
	SharedRequestedTransactions() *SharedRequestedTransactions
	Broadcast(message appmessage.Message) error
	OnTransactionAddedToMempool()
}

type handleRelayedTransactionsFlow struct {
	TransactionsRelayContext
	incomingRoute, outgoingRoute *router.Route
	invsQueue                    []*appmessage.MsgInvTransaction
}

// HandleRelayedTransactions listens to appmessage.MsgInvTransaction messages, requests their corresponding transactions if they
// are missing, adds them to the mempool and propagates them to the rest of the network.
func HandleRelayedTransactions(context TransactionsRelayContext, incomingRoute *router.Route, outgoingRoute *router.Route) error {
	flow := &handleRelayedTransactionsFlow{
		TransactionsRelayContext: context,
		incomingRoute:            incomingRoute,
		outgoingRoute:            outgoingRoute,
		invsQueue:                make([]*appmessage.MsgInvTransaction, 0),
	}
	return flow.start()
}

func (flow *handleRelayedTransactionsFlow) start() error {
	for {
		inv, err := flow.readInv()
		if err != nil {
			return err
		}

		requestedIDs, err := flow.requestInvTransactions(inv)
		if err != nil {
			return err
		}

		err = flow.receiveTransactions(requestedIDs)
		if err != nil {
			return err
		}
	}
}

func (flow *handleRelayedTransactionsFlow) requestInvTransactions(
	inv *appmessage.MsgInvTransaction) (requestedIDs []*externalapi.DomainTransactionID, err error) {

	idsToRequest := make([]*externalapi.DomainTransactionID, 0, len(inv.TxIDs))
	for _, txID := range inv.TxIDs {
		if flow.isKnownTransaction(txID) {
			continue
		}
		exists := flow.SharedRequestedTransactions().addIfNotExists(txID)
		if exists {
			continue
		}
		idsToRequest = append(idsToRequest, txID)
	}

	if len(idsToRequest) == 0 {
		return idsToRequest, nil
	}

	msgGetTransactions := appmessage.NewMsgRequestTransactions(idsToRequest)
	err = flow.outgoingRoute.Enqueue(msgGetTransactions)
	if err != nil {
		flow.SharedRequestedTransactions().removeMany(idsToRequest)
		return nil, err
	}
	return idsToRequest, nil
}

func (flow *handleRelayedTransactionsFlow) isKnownTransaction(txID *externalapi.DomainTransactionID) bool {
	// Ask the transaction memory pool if the transaction is known
	// to it in any form (main pool or orphan).
	if _, ok := flow.Domain().MiningManager().GetTransaction(txID); ok {
		return true
	}

	return false
}

func (flow *handleRelayedTransactionsFlow) readInv() (*appmessage.MsgInvTransaction, error) {
	if len(flow.invsQueue) > 0 {
		var inv *appmessage.MsgInvTransaction
		inv, flow.invsQueue = flow.invsQueue[0], flow.invsQueue[1:]
		return inv, nil
	}

	msg, err := flow.incomingRoute.Dequeue()
	if err != nil {
		return nil, err
	}

	inv, ok := msg.(*appmessage.MsgInvTransaction)
	if !ok {
		return nil, protocolerrors.Errorf(true, "unexpected %s message in the block relay flow while "+
			"expecting an inv message", msg.Command())
	}
	return inv, nil
}

func (flow *handleRelayedTransactionsFlow) broadcastAcceptedTransactions(acceptedTxIDs []*externalapi.DomainTransactionID) error {
	inv := appmessage.NewMsgInvTransaction(acceptedTxIDs)
	return flow.Broadcast(inv)
}

// readMsgTxOrNotFound returns the next msgTx or msgTransactionNotFound in incomingRoute,
// returning only one of the message types at a time.
//
// and populates invsQueue with any inv messages that meanwhile arrive.
func (flow *handleRelayedTransactionsFlow) readMsgTxOrNotFound() (
	msgTx *appmessage.MsgTx, msgNotFound *appmessage.MsgTransactionNotFound, err error) {

	for {
		message, err := flow.incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
		if err != nil {
			return nil, nil, err
		}

		switch message := message.(type) {
		case *appmessage.MsgInvTransaction:
			flow.invsQueue = append(flow.invsQueue, message)
		case *appmessage.MsgTx:
			return message, nil, nil
		case *appmessage.MsgTransactionNotFound:
			return nil, message, nil
		default:
			return nil, nil, errors.Errorf("unexpected message %s", message.Command())
		}
	}
}

func (flow *handleRelayedTransactionsFlow) receiveTransactions(requestedTransactions []*externalapi.DomainTransactionID) error {
	// In case the function returns earlier than expected, we want to make sure sharedRequestedTransactions is
	// clean from any pending transactions.
	defer flow.SharedRequestedTransactions().removeMany(requestedTransactions)
	for _, expectedID := range requestedTransactions {
		msgTx, msgTxNotFound, err := flow.readMsgTxOrNotFound()
		if err != nil {
			return err
		}
		if msgTxNotFound != nil {
			if !msgTxNotFound.ID.Equal(expectedID) {
				return protocolerrors.Errorf(true, "expected transaction %s, but got %s",
					expectedID, msgTxNotFound.ID)
			}

			continue
		}
		tx := appmessage.MsgTxToDomainTransaction(msgTx)
		txID := consensushashing.TransactionID(tx)
		if !txID.Equal(expectedID) {
			return protocolerrors.Errorf(true, "expected transaction %s, but got %s",
				expectedID, txID)
		}

		err = flow.Domain().MiningManager().ValidateAndInsertTransaction(tx, true)
		if err != nil {
			ruleErr := &mempool_old.RuleError{}
			if !errors.As(err, ruleErr) {
				return errors.Wrapf(err, "failed to process transaction %s", txID)
			}

			shouldBan := false
			if txRuleErr := (&mempool_old.TxRuleError{}); errors.As(ruleErr.Err, txRuleErr) {
				if txRuleErr.RejectCode == mempool_old.RejectInvalid {
					shouldBan = true
				}
			}

			if !shouldBan {
				continue
			}

			return protocolerrors.Errorf(true, "rejected transaction %s: %s", txID, ruleErr)
		}
		err = flow.broadcastAcceptedTransactions([]*externalapi.DomainTransactionID{txID})
		if err != nil {
			return err
		}
		flow.OnTransactionAddedToMempool()
	}
	return nil
}
