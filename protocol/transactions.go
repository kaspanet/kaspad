package protocol

import (
	"github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
	"time"
)

func (m *Manager) AddTransaction(tx *util.Tx) error {
	acceptedTxs, err := m.txPool.ProcessTransaction(tx, false, 0)
	if err != nil {
		return err
	}

	if len(acceptedTxs) > 0 {
		panic(errors.New("got accepted transaction when no orphans were allowed"))
	}

	m.transactionsToRebroadcast[*tx.ID()] = tx
	// TODO(libp2p) Implement inv type to relay txs and broadcast.
	return m.netAdapter.Broadcast(peer.ReadyPeerIDs(), tx.MsgTx())
}

func (m *Manager) updateTransactionsToRebroadcast(block *util.Block) {
	// Note: if the block is red, its transactions won't be rebroadcasted
	// anymore, although they are not included in the UTXO set.
	// This is probably ok, since red blocks are quite rare.
	for _, tx := range block.Transactions() {
		delete(m.transactionsToRebroadcast, *tx.ID())
	}
}

func (m *Manager) maybeRebroadcastTransactions() {
	if len(m.transactionsToRebroadcast) == 0 {
		return
	}

	const rebroadcastInterval = 30 * time.Second
	if time.Since(m.lastRebroadcastTime) > rebroadcastInterval {
		return
	}

	// TODO(libp2p) Implement inv type to relay txs and broadcast.
}
