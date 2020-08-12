package relaytransactions

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/domainmessage"
	"github.com/kaspanet/kaspad/mempool"
	"github.com/kaspanet/kaspad/netadapter"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/common"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/pkg/errors"
)

// TransactionsRelayContext is the interface for the context needed for the
// HandleRelayedTransactions and HandleRequestedTransactions flows.
type TransactionsRelayContext interface {
	NetAdapter() *netadapter.NetAdapter
	DAG() *blockdag.BlockDAG
	SharedRequestedTransactions() *SharedRequestedTransactions
	TxPool() *mempool.TxPool
	Broadcast(message domainmessage.Message) error
}

type handleRelayedTransactionsFlow struct {
	TransactionsRelayContext
	incomingRoute, outgoingRoute *router.Route
	invsQueue                    []*domainmessage.MsgInvTransaction
}

// HandleRelayedTransactions listens to domainmessage.MsgInvTransaction messages, requests their corresponding transactions if they
// are missing, adds them to the mempool and propagates them to the rest of the network.
func HandleRelayedTransactions(context TransactionsRelayContext, incomingRoute *router.Route, outgoingRoute *router.Route) error {
	flow := &handleRelayedTransactionsFlow{
		TransactionsRelayContext: context,
		incomingRoute:            incomingRoute,
		outgoingRoute:            outgoingRoute,
		invsQueue:                make([]*domainmessage.MsgInvTransaction, 0),
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
	inv *domainmessage.MsgInvTransaction) (requestedIDs []*daghash.TxID, err error) {

	idsToRequest := make([]*daghash.TxID, 0, len(inv.TxIDs))
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

	msgGetTransactions := domainmessage.NewMsgRequestTransactions(idsToRequest)
	err = flow.outgoingRoute.Enqueue(msgGetTransactions)
	if err != nil {
		flow.SharedRequestedTransactions().removeMany(idsToRequest)
		return nil, err
	}
	return idsToRequest, nil
}

func (flow *handleRelayedTransactionsFlow) isKnownTransaction(txID *daghash.TxID) bool {
	// Ask the transaction memory pool if the transaction is known
	// to it in any form (main pool or orphan).
	if flow.TxPool().HaveTransaction(txID) {
		return true
	}

	// Check if the transaction exists from the point of view of the
	// DAG's virtual block. Note that this is only a best effort
	// since it is expensive to check existence of every output and
	// the only purpose of this check is to avoid downloading
	// already known transactions. Only the first two outputs are
	// checked because the vast majority of transactions consist of
	// two outputs where one is some form of "pay-to-somebody-else"
	// and the other is a change output.
	prevOut := domainmessage.Outpoint{TxID: *txID}
	for i := uint32(0); i < 2; i++ {
		prevOut.Index = i
		_, ok := flow.DAG().GetUTXOEntry(prevOut)
		if ok {
			return true
		}
	}
	return false
}

func (flow *handleRelayedTransactionsFlow) readInv() (*domainmessage.MsgInvTransaction, error) {
	if len(flow.invsQueue) > 0 {
		var inv *domainmessage.MsgInvTransaction
		inv, flow.invsQueue = flow.invsQueue[0], flow.invsQueue[1:]
		return inv, nil
	}

	msg, err := flow.incomingRoute.Dequeue()
	if err != nil {
		return nil, err
	}

	inv, ok := msg.(*domainmessage.MsgInvTransaction)
	if !ok {
		return nil, protocolerrors.Errorf(true, "unexpected %s message in the block relay flow while "+
			"expecting an inv message", msg.Command())
	}
	return inv, nil
}

func (flow *handleRelayedTransactionsFlow) broadcastAcceptedTransactions(acceptedTxs []*mempool.TxDesc) error {
	// TODO(libp2p) Add mechanism to avoid sending to other peers invs that are known to them (e.g. mruinvmap)
	// TODO(libp2p) Consider broadcasting in bulks
	idsToBroadcast := make([]*daghash.TxID, len(acceptedTxs))
	for i, tx := range acceptedTxs {
		idsToBroadcast[i] = tx.Tx.ID()
	}
	inv := domainmessage.NewMsgInvTransaction(idsToBroadcast)
	return flow.Broadcast(inv)
}

// readMsgTxOrNotFound returns the next msgTx or msgTransactionNotFound in incomingRoute,
// returning only one of the message types at a time.
//
// and populates invsQueue with any inv messages that meanwhile arrive.
func (flow *handleRelayedTransactionsFlow) readMsgTxOrNotFound() (
	msgTx *domainmessage.MsgTx, msgNotFound *domainmessage.MsgTransactionNotFound, err error) {

	for {
		message, err := flow.incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
		if err != nil {
			return nil, nil, err
		}

		switch message := message.(type) {
		case *domainmessage.MsgInvTransaction:
			flow.invsQueue = append(flow.invsQueue, message)
		case *domainmessage.MsgTx:
			return message, nil, nil
		case *domainmessage.MsgTransactionNotFound:
			return nil, message, nil
		default:
			return nil, nil, errors.Errorf("unexpected message %s", message.Command())
		}
	}
}

func (flow *handleRelayedTransactionsFlow) receiveTransactions(requestedTransactions []*daghash.TxID) error {
	// In case the function returns earlier than expected, we want to make sure sharedRequestedTransactions is
	// clean from any pending transactions.
	defer flow.SharedRequestedTransactions().removeMany(requestedTransactions)
	for _, expectedID := range requestedTransactions {
		msgTx, msgTxNotFound, err := flow.readMsgTxOrNotFound()
		if err != nil {
			return err
		}
		if msgTxNotFound != nil {
			if !msgTxNotFound.ID.IsEqual(expectedID) {
				return protocolerrors.Errorf(true, "expected transaction %s, but got %s",
					expectedID, msgTxNotFound.ID)
			}

			continue
		}
		tx := util.NewTx(msgTx)
		if !tx.ID().IsEqual(expectedID) {
			return protocolerrors.Errorf(true, "expected transaction %s, but got %s",
				expectedID, tx.ID())
		}

		acceptedTxs, err := flow.TxPool().ProcessTransaction(tx, true, 0) // TODO(libp2p) Use the peer ID for the mempool tag
		if err != nil {
			ruleErr := &mempool.RuleError{}
			if !errors.As(err, ruleErr) {
				return errors.Wrapf(err, "failed to process transaction %s", tx.ID())
			}

			shouldBan := false
			if txRuleErr := (&mempool.TxRuleError{}); errors.As(ruleErr.Err, txRuleErr) {
				if txRuleErr.RejectCode == mempool.RejectInvalid {
					shouldBan = true
				}
			} else if dagRuleErr := (&blockdag.RuleError{}); errors.As(ruleErr.Err, dagRuleErr) {
				shouldBan = true
			}

			if !shouldBan {
				continue
			}

			return protocolerrors.Errorf(true, "rejected transaction %s", tx.ID())
		}
		err = flow.broadcastAcceptedTransactions(acceptedTxs)
		if err != nil {
			return err
		}
		// TODO(libp2p) Notify transactionsAcceptedToMempool to RPC
	}
	return nil
}
