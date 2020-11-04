package flowcontext

import (
	"sync/atomic"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashserialization"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/flows/blockrelay"
)

// OnNewBlock updates the mempool after a new block arrival, and
// relays newly unorphaned transactions and possibly rebroadcast
// manually added transactions when not in IBD.
func (f *FlowContext) OnNewBlock(block *externalapi.DomainBlock) error {
	f.Domain().HandleNewBlockTransactions(block.Transactions)

	if f.onBlockAddedToDAGHandler != nil {
		err := f.onBlockAddedToDAGHandler(block)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *FlowContext) broadcastTransactionsAfterBlockAdded(
	block *externalapi.DomainBlock, transactionsAcceptedToMempool []*externalapi.DomainTransaction) error {

	f.updateTransactionsToRebroadcast(block)

	// Don't relay transactions when in IBD.
	if atomic.LoadUint32(&f.isInIBD) != 0 {
		return nil
	}

	var txIDsToRebroadcast []*externalapi.DomainTransactionID
	if f.shouldRebroadcastTransactions() {
		txIDsToRebroadcast = f.txIDsToRebroadcast()
	}

	txIDsToBroadcast := make([]*externalapi.DomainTransactionID, len(transactionsAcceptedToMempool)+len(txIDsToRebroadcast))
	for i, tx := range transactionsAcceptedToMempool {
		txIDsToBroadcast[i] = hashserialization.TransactionID(tx)
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
	err := f.Domain().ValidateAndInsertBlock(block)
	if err != nil {
		return err
	}
	err = f.OnNewBlock(block)
	if err != nil {
		return err
	}
	return f.Broadcast(appmessage.NewMsgInvBlock(hashserialization.BlockHash(block)))
}
