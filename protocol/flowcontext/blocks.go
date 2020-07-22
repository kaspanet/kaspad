package flowcontext

import (
	"github.com/kaspanet/kaspad/protocol/flows/blockrelay"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
	"sync/atomic"
)

// OnNewBlock updates the mempool after a new block arrival, and
// relays newly unorphaned transactions and possibly rebroadcast
// manually added transactions when not in IBD.
func (f *FlowContext) OnNewBlock(block *util.Block) error {
	transactionsAcceptedToMempool, err := f.txPool.HandleNewBlock(block)
	if err != nil {
		return err
	}
	// TODO(libp2p) Notify transactionsAcceptedToMempool to RPC

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

	copy(txIDsToBroadcast[len(transactionsAcceptedToMempool):], txIDsToBroadcast)
	txIDsToBroadcast = txIDsToBroadcast[:wire.MaxInvPerTxInvMsg]
	inv := wire.NewMsgTxInv(txIDsToBroadcast)
	return f.Broadcast(inv)
}

// SharedRequestedBlocks returns a *blockrelay.SharedRequestedBlocks for sharing
// data about requested blocks between different peers.
func (f *FlowContext) SharedRequestedBlocks() *blockrelay.SharedRequestedBlocks {
	return f.sharedRequestedBlocks
}

// AddBlock adds the given block to the DAG and propagates it.
func (f *FlowContext) AddBlock(block *util.Block) error {
	// TODO(libp2p): unimplemented
	panic("unimplemented")
}
