package protocol

import (
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
	"sync/atomic"
)

// OnNewBlock updates the mempool after a new block arrival, and
// relays newly unorphaned transactions and possibly rebroadcast
// manually added transactions when not in IBD.
// TODO(libp2p) Call this function from IBD as well.
func (m *Manager) OnNewBlock(block *util.Block) error {
	transactionsAcceptedToMempool, err := m.txPool.HandleNewBlock(block)
	if err != nil {
		return err
	}
	// TODO(libp2p) Notify transactionsAcceptedToMempool to RPC

	m.updateTransactionsToRebroadcast(block)

	// Don't relay transactions when in IBD.
	if atomic.LoadUint32(&m.isInIBD) != 0 {
		return nil
	}

	var txIDsToRebroadcast []*daghash.TxID
	if m.shouldRebroadcastTransactions() {
		txIDsToRebroadcast = m.txIDsToRebroadcast()
	}

	txIDsToBroadcast := make([]*daghash.TxID, len(transactionsAcceptedToMempool)+len(txIDsToRebroadcast))
	for i, tx := range transactionsAcceptedToMempool {
		txIDsToBroadcast[i] = tx.ID()
	}

	copy(txIDsToBroadcast[len(transactionsAcceptedToMempool):], txIDsToBroadcast)
	txIDsToBroadcast = txIDsToBroadcast[:wire.MaxInvPerTxInvMsg]
	inv := wire.NewMsgTxInv(txIDsToBroadcast)
	return m.netAdapter.Broadcast(m.peers.ReadyPeerIDs(), inv)
}

// AddBlock adds the given block to the DAG and propagates it.
func (m *Manager) AddBlock(block *util.Block) error {
	// TODO(libp2p): unimplemented
	panic("unimplemented")
}
