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

// NewBlockHandler is a function that is to be
// called when a new block is successfully processed.
type NewBlockHandler func(block *util.Block) error

// HandleRelayedTransactions listens to wire.MsgInvTransaction messages, requests their corresponding transactions if they
// are missing, adds them to the mempool and propagates them to the rest of the network.
func HandleRelayedTransactions(incomingRoute *router.Route, outgoingRoute *router.Route,
	netAdapter *netadapter.NetAdapter, dag *blockdag.BlockDAG, txPool *mempool.TxPool,
	sharedRequestedTransactions *SharedRequestedTransactions) error {

	invsQueue := make([]*wire.MsgInvTransaction, 0)
	for {
		inv, err := readInv(incomingRoute, &invsQueue)
		if err != nil {
			return err
		}

		requestedIDs, err := requestInvTransactions(outgoingRoute, txPool, dag, sharedRequestedTransactions, inv)
		if err != nil {
			return err
		}

		err = receiveTransactions(requestedIDs, incomingRoute, &invsQueue, txPool, netAdapter,
			sharedRequestedTransactions)
		if err != nil {
			return err
		}
	}
}

func requestInvTransactions(outgoingRoute *router.Route, txPool *mempool.TxPool, dag *blockdag.BlockDAG,
	sharedRequestedTransactions *SharedRequestedTransactions, inv *wire.MsgInvTransaction) (requestedIDs []*daghash.TxID,
	err error) {

	idsToRequest := make([]*daghash.TxID, 0, len(inv.TxIDS))
	for _, txID := range inv.TxIDS {
		if isKnownTransaction(txPool, dag, txID) {
			continue
		}
		exists := sharedRequestedTransactions.addIfNotExists(txID)
		if exists {
			continue
		}
		idsToRequest = append(idsToRequest, txID)
	}

	if len(idsToRequest) == 0 {
		return idsToRequest, nil
	}

	msgGetTransactions := wire.NewMsgGetTransactions(idsToRequest)
	isOpen := outgoingRoute.Enqueue(msgGetTransactions)
	if !isOpen {
		sharedRequestedTransactions.removeMany(idsToRequest)
		return nil, errors.WithStack(common.ErrRouteClosed)
	}
	return idsToRequest, nil
}

func isKnownTransaction(txPool *mempool.TxPool, dag *blockdag.BlockDAG, txID *daghash.TxID) bool {
	// Ask the transaction memory pool if the transaction is known
	// to it in any form (main pool or orphan).
	if txPool.HaveTransaction(txID) {
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
		_, ok := dag.GetUTXOEntry(prevOut)
		if ok {
			return true
		}
	}
	return false
}

func readInv(incomingRoute *router.Route, invsQueue *[]*wire.MsgInvTransaction) (
	inv *wire.MsgInvTransaction, err error) {

	if len(*invsQueue) > 0 {
		inv, *invsQueue = (*invsQueue)[0], (*invsQueue)[1:]
		return inv, nil
	}

	msg, isOpen := incomingRoute.Dequeue()
	if !isOpen {
		return nil, errors.WithStack(common.ErrRouteClosed)
	}

	inv, ok := msg.(*wire.MsgInvTransaction)
	if !ok {
		return nil, protocolerrors.Errorf(true, "unexpected %s message in the block relay flow while "+
			"expecting an inv message", msg.Command())
	}
	return inv, nil
}

func broadcastAcceptedTransactions(netAdapter *netadapter.NetAdapter, acceptedTxs []*mempool.TxDesc) error {
	// TODO(libp2p) Add mechanism to avoid sending to other peers invs that are known to them (e.g. mruinvmap)
	// TODO(libp2p) Consider broadcasting in bulks
	idsToBroadcast := make([]*daghash.TxID, len(acceptedTxs))
	for i, tx := range acceptedTxs {
		idsToBroadcast[i] = tx.Tx.ID()
	}
	inv := wire.NewMsgTxInv(idsToBroadcast)
	return netAdapter.Broadcast(peerpkg.ReadyPeerIDs(), inv)
}

// readMsgTx returns the next msgTx in incomingRoute, and populates invsQueue with any inv messages that meanwhile arrive.
//
// Note: this function assumes msgChan can contain only wire.MsgInvTransaction and wire.MsgBlock messages.
func readMsgTx(incomingRoute *router.Route, invsQueue *[]*wire.MsgInvTransaction) (
	msgTx *wire.MsgTx, err error) {

	for {
		message, isOpen, err := incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
		if err != nil {
			return nil, err
		}
		if !isOpen {
			return nil, errors.WithStack(common.ErrRouteClosed)
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

func receiveTransactions(requestedTransactions []*daghash.TxID, incomingRoute *router.Route,
	invsQueue *[]*wire.MsgInvTransaction, txPool *mempool.TxPool, netAdapter *netadapter.NetAdapter,
	sharedRequestedTransactions *SharedRequestedTransactions) error {

	// In case the function returns earlier than expected, we want to make sure sharedRequestedTransactions is
	// clean from any pending transactions.
	defer sharedRequestedTransactions.removeMany(requestedTransactions)
	for _, expectedID := range requestedTransactions {
		msgTx, err := readMsgTx(incomingRoute, invsQueue)
		if err != nil {
			return err
		}
		tx := util.NewTx(msgTx)
		if !tx.ID().IsEqual(expectedID) {
			return protocolerrors.Errorf(true, "expected transaction %s", expectedID)
		}

		acceptedTxs, err := txPool.ProcessTransaction(tx, true, 0) // TODO(libp2p) Use the peer ID for the mempool tag
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
		err = broadcastAcceptedTransactions(netAdapter, acceptedTxs)
		if err != nil {
			panic(err)
		}
		// TODO(libp2p) Notify transactionsAcceptedToMempool to RPC
	}
	return nil
}
