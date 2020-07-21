package relaytransactions

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/mempool"
	"github.com/kaspanet/kaspad/netadapter"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/common"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

type RelayedTransactionsContext interface {
	NetAdapter() *netadapter.NetAdapter
	DAG() *blockdag.BlockDAG
	SharedRequestedTransactions() *SharedRequestedTransactions
	TxPool() *mempool.TxPool
}

// HandleRelayedTransactions listens to wire.MsgInvTransaction messages, requests their corresponding transactions if they
// are missing, adds them to the mempool and propagates them to the rest of the network.
func HandleRelayedTransactions(context RelayedTransactionsContext, incomingRoute *router.Route, outgoingRoute *router.Route) error {

	invsQueue := make([]*wire.MsgInvTransaction, 0)
	for {
		inv, err := readInv(incomingRoute, &invsQueue)
		if err != nil {
			return err
		}

		requestedIDs, err := requestInvTransactions(context, outgoingRoute, inv)
		if err != nil {
			return err
		}

		err = receiveTransactions(context, requestedIDs, incomingRoute, &invsQueue)
		if err != nil {
			return err
		}
	}
}

func requestInvTransactions(context RelayedTransactionsContext, outgoingRoute *router.Route,
	inv *wire.MsgInvTransaction) (requestedIDs []*daghash.TxID, err error) {

	idsToRequest := make([]*daghash.TxID, 0, len(inv.TxIDS))
	for _, txID := range inv.TxIDS {
		if isKnownTransaction(context, txID) {
			continue
		}
		exists := context.SharedRequestedTransactions().addIfNotExists(txID)
		if exists {
			continue
		}
		idsToRequest = append(idsToRequest, txID)
	}

	if len(idsToRequest) == 0 {
		return idsToRequest, nil
	}

	msgGetTransactions := wire.NewMsgGetTransactions(idsToRequest)
	err = outgoingRoute.Enqueue(msgGetTransactions)
	if err != nil {
		context.SharedRequestedTransactions().removeMany(idsToRequest)
		return nil, err
	}
	return idsToRequest, nil
}

func isKnownTransaction(context RelayedTransactionsContext, txID *daghash.TxID) bool {
	// Ask the transaction memory pool if the transaction is known
	// to it in any form (main pool or orphan).
	if context.TxPool().HaveTransaction(txID) {
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
		_, ok := context.DAG().GetUTXOEntry(prevOut)
		if ok {
			return true
		}
	}
	return false
}

func readInv(incomingRoute *router.Route, invsQueue *[]*wire.MsgInvTransaction) (*wire.MsgInvTransaction, error) {

	if len(*invsQueue) > 0 {
		var inv *wire.MsgInvTransaction
		inv, *invsQueue = (*invsQueue)[0], (*invsQueue)[1:]
		return inv, nil
	}

	msg, err := incomingRoute.Dequeue()
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

func broadcastAcceptedTransactions(context RelayedTransactionsContext, acceptedTxs []*mempool.TxDesc) error {
	// TODO(libp2p) Add mechanism to avoid sending to other peers invs that are known to them (e.g. mruinvmap)
	// TODO(libp2p) Consider broadcasting in bulks
	idsToBroadcast := make([]*daghash.TxID, len(acceptedTxs))
	for i, tx := range acceptedTxs {
		idsToBroadcast[i] = tx.Tx.ID()
	}
	inv := wire.NewMsgTxInv(idsToBroadcast)
	return context.NetAdapter().Broadcast(peerpkg.ReadyPeerIDs(), inv)
}

// readMsgTx returns the next msgTx in incomingRoute, and populates invsQueue with any inv messages that meanwhile arrive.
//
// Note: this function assumes msgChan can contain only wire.MsgInvTransaction and wire.MsgBlock messages.
func readMsgTx(incomingRoute *router.Route, invsQueue *[]*wire.MsgInvTransaction) (
	msgTx *wire.MsgTx, err error) {

	for {
		message, err := incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
		if err != nil {
			return nil, err
		}

		switch message := message.(type) {
		case *wire.MsgInvTransaction:
			*invsQueue = append(*invsQueue, message)
		case *wire.MsgTx:
			return message, nil
		default:
			panic(errors.Errorf("unexpected message %s", message.Command()))
		}
	}
}

func receiveTransactions(context RelayedTransactionsContext, requestedTransactions []*daghash.TxID, incomingRoute *router.Route,
	invsQueue *[]*wire.MsgInvTransaction) error {

	// In case the function returns earlier than expected, we want to make sure sharedRequestedTransactions is
	// clean from any pending transactions.
	defer context.SharedRequestedTransactions().removeMany(requestedTransactions)
	for _, expectedID := range requestedTransactions {
		msgTx, err := readMsgTx(incomingRoute, invsQueue)
		if err != nil {
			return err
		}
		tx := util.NewTx(msgTx)
		if !tx.ID().IsEqual(expectedID) {
			return protocolerrors.Errorf(true, "expected transaction %s", expectedID)
		}

		acceptedTxs, err := context.TxPool().ProcessTransaction(tx, true, 0) // TODO(libp2p) Use the peer ID for the mempool tag
		if err != nil {
			// When the error is a rule error, it means the transaction was
			// simply rejected as opposed to something actually going wrong,
			// so log it as such. Otherwise, something really did go wrong,
			// so panic.
			ruleErr := &mempool.RuleError{}
			if !errors.As(err, ruleErr) {
				panic(errors.Wrapf(err, "failed to process transaction %s", tx.ID()))
			}

			shouldBan := false
			if txRuleErr := (&mempool.TxRuleError{}); errors.As(ruleErr.Err, txRuleErr) {
				if txRuleErr.RejectCode == wire.RejectInvalid {
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
		err = broadcastAcceptedTransactions(context, acceptedTxs)
		if err != nil {
			panic(err)
		}
		// TODO(libp2p) Notify transactionsAcceptedToMempool to RPC
	}
	return nil
}
