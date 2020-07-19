package protocol

import (
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

func (m *Manager) OnNewBlock(block *util.Block) error {
	acceptedTxs, err := m.txPool.HandleNewBlock(block)
	if err != nil {
		return err
	}

	m.updateTransactionsToRebroadcast(block)

	var txIDsToRebroadcast []*daghash.TxID
	if m.shouldRebroadcastTransactions() {
		txIDsToRebroadcast = m.txIDsToRebroadcast()
	}

	txIDsToBroadcast := make([]*daghash.TxID, len(acceptedTxs)+len(txIDsToRebroadcast))
	for i, tx := range acceptedTxs {
		txIDsToBroadcast[i] = tx.ID()
	}

	copy(txIDsToBroadcast[len(acceptedTxs):], txIDsToBroadcast)
	txIDsToBroadcast = txIDsToBroadcast[:wire.MaxInvPerTxInvMsg]
	inv := wire.NewMsgTxInv(txIDsToBroadcast)
	return m.netAdapter.Broadcast(peerpkg.ReadyPeerIDs(), inv)
}
