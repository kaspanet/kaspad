package protocol

import (
	"github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
	"time"
)

func (m *Manager) AddTransaction(tx *util.Tx) error {
	m.transactionsToRebroadcastLock.Lock()
	defer m.transactionsToRebroadcastLock.Unlock()
	acceptedTxs, err := m.txPool.ProcessTransaction(tx, false, 0)
	if err != nil {
		return err
	}

	if len(acceptedTxs) > 1 {
		panic(errors.New("got more than one accepted transactions when no orphans were allowed"))
	}

	m.transactionsToRebroadcast[*tx.ID()] = tx
	inv := wire.NewMsgTxInv([]*daghash.TxID{tx.ID()})
	return m.netAdapter.Broadcast(peer.ReadyPeerIDs(), inv)
}

func (m *Manager) updateTransactionsToRebroadcast(block *util.Block) {
	m.transactionsToRebroadcastLock.Lock()
	defer m.transactionsToRebroadcastLock.Unlock()
	// Note: if the block is red, its transactions won't be rebroadcasted
	// anymore, although they are not included in the UTXO set.
	// This is probably ok, since red blocks are quite rare.
	for _, tx := range block.Transactions() {
		delete(m.transactionsToRebroadcast, *tx.ID())
	}
}

func (m *Manager) shouldRebroadcastTransactions() bool {
	const rebroadcastInterval = 30 * time.Second
	return time.Since(m.lastRebroadcastTime) > rebroadcastInterval
}

func (m *Manager) txIDsToRebroadcast() []*daghash.TxID {
	m.transactionsToRebroadcastLock.Lock()
	defer m.transactionsToRebroadcastLock.Unlock()

	txIDs := make([]*daghash.TxID, len(m.transactionsToRebroadcast))
	i := 0
	for _, tx := range m.transactionsToRebroadcast {
		txIDs[i] = tx.ID()
		i++
	}
	return txIDs
}
