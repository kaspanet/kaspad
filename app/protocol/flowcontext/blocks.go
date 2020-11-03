package flowcontext

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"sync/atomic"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/flows/blockrelay"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
)

// OnNewBlock updates the mempool after a new block arrival, and
// relays newly unorphaned transactions and possibly rebroadcast
// manually added transactions when not in IBD.
func (f *FlowContext) OnNewBlock(block *util.Block) error {
	transactionsAcceptedToMempool, err := f.txPool.HandleNewBlock(block)
	if err != nil {
		return err
	}
	if f.onBlockAddedToDAGHandler != nil {
		err := f.onBlockAddedToDAGHandler(block)
		if err != nil {
			return err
		}
	}

	return f.broadcastTransactionsAfterBlockAdded(block, transactionsAcceptedToMempool)
}

func (f *FlowContext) broadcastTransactionsAfterBlockAdded(block *util.Block, transactionsAcceptedToMempool []*util.Tx) error {
	f.updateTransactionsToRebroadcast(block)

	// Don't relay transactions when in IBD.
	if atomic.LoadUint32(&f.isInIBD) != 0 {
		return nil
	}

	var txIDsToRebroadcast []*daghash.TxID
	if f.shouldRebroadcastTransactions() {
		txIDsToRebroadcast = f.txIDsToRebroadcast()
	}

	txIDsToBroadcast := make([]*daghash.TxID, len(transactionsAcceptedToMempool)+len(txIDsToRebroadcast))
	for i, tx := range transactionsAcceptedToMempool {
		txIDsToBroadcast[i] = tx.ID()
	}
	offset := len(transactionsAcceptedToMempool)
	for i, txID := range txIDsToRebroadcast {
		txIDsToBroadcast[offset+i] = txID
	}

	if len(txIDsToBroadcast) == 0 {
		return nil
	}
	if len(txIDsToBroadcast) > appmessage.MaxInvPerTxInvMsg {
		txIDsToBroadcast = txIDsToBroadcast[:appmessage.MaxInvPerTxInvMsg]
	}
	inv := appmessage.NewMsgInvTransaction(txIDsToBroadcast)
	return f.Broadcast(inv)
}

// SharedRequestedBlocks returns a *blockrelay.SharedRequestedBlocks for sharing
// data about requested blocks between different peers.
func (f *FlowContext) SharedRequestedBlocks() *blockrelay.SharedRequestedBlocks {
	return f.sharedRequestedBlocks
}

// AddBlock adds the given block to the DAG and propagates it.
func (f *FlowContext) AddBlock(block *externalapi.DomainBlock) error {
	_, _, err := f.Domain().ValidateAndInsertBlock(block, false)
	if err != nil {
		return err
	}
	err = f.OnNewBlock(block)
	if err != nil {
		return err
	}
	return f.Broadcast(appmessage.NewMsgInvBlock(block.Hash()))
}
