package relaytransactions

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/mempool"
	"github.com/kaspanet/kaspad/netadapter"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/common"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

// RelayedTransactionsContext is the interface for the context needed for the HandleRelayedTransactions flow.
type RelayedTransactionsContext interface {
	NetAdapter() *netadapter.NetAdapter
	DAG() *blockdag.BlockDAG
	SharedRequestedTransactions() *SharedRequestedTransactions
	TxPool() *mempool.TxPool
	Broadcast(message wire.Message) error
}

type handleRelayedTransactionsFlow struct {
	RelayedTransactionsContext
	incomingRoute, outgoingRoute *router.Route
	invsQueue                    []*wire.MsgInvTransaction
}

// HandleRelayedTransactions listens to wire.MsgInvTransaction messages, requests their corresponding transactions if they
// are missing, adds them to the mempool and propagates them to the rest of the network.
func HandleRelayedTransactions(context RelayedTransactionsContext, incomingRoute *router.Route, outgoingRoute *router.Route) error {
	flow := &handleRelayedTransactionsFlow{
		RelayedTransactionsContext: context,
		incomingRoute:              incomingRoute,
		outgoingRoute:              outgoingRoute,
		invsQueue:                  make([]*wire.MsgInvTransaction, 0),
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
	inv *wire.MsgInvTransaction) (requestedIDs []*daghash.TxID, err error) {

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

	msgGetTransactions := wire.NewMsgGetTransactions(idsToRequest)
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
	prevOut := wire.Outpoint{TxID: *txID}
	for i := uint32(0); i < 2; i++ {
		prevOut.Index = i
		_, ok := flow.DAG().GetUTXOEntry(prevOut)
		if ok {
			return true
		}
	}
	return false
}

func (flow *handleRelayedTransactionsFlow) readInv() (*wire.MsgInvTransaction, error) {

	if len(flow.invsQueue) > 0 {
		var inv *wire.MsgInvTransaction
		inv, flow.invsQueue = flow.invsQueue[0], flow.invsQueue[1:]
		return inv, nil
	}

	msg, err := flow.incomingRoute.Dequeue()
	if err != nil {
		return nil, err
	}

	inv, ok := msg.(*wire.MsgInvTransaction)
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
	inv := wire.NewMsgInvTransaction(idsToBroadcast)
	return flow.Broadcast(inv)
}

// readMsgTx returns the next msgTx in incomingRoute, and populates invsQueue with any inv messages that meanwhile arrive.
//
// Note: this function assumes msgChan can contain only wire.MsgInvTransaction and wire.MsgBlock messages.
func (flow *handleRelayedTransactionsFlow) readMsgTx() (
	msgTx *wire.MsgTx, err error) {

	for {
		message, err := flow.incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
		if err != nil {
			return nil, err
		}

		switch message := message.(type) {
		case *wire.MsgInvTransaction:
			flow.invsQueue = append(flow.invsQueue, message)
		case *wire.MsgTx:
			return message, nil
		default:
			return nil, errors.Errorf("unexpected message %s", message.Command())
		}
	}
}

func (flow *handleRelayedTransactionsFlow) receiveTransactions(requestedTransactions []*daghash.TxID) error {

	// In case the function returns earlier than expected, we want to make sure sharedRequestedTransactions is
	// clean from any pending transactions.
	defer flow.SharedRequestedTransactions().removeMany(requestedTransactions)
	for _, expectedID := range requestedTransactions {
		msgTx, err := flow.readMsgTx()
		if err != nil {
			return err
		}
		tx := util.NewTx(msgTx)
		if !tx.ID().IsEqual(expectedID) {
			return protocolerrors.Errorf(true, "expected transaction %s", expectedID)
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
