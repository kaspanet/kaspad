package protocol

import (
	"github.com/kaspanet/kaspad/util"
)

func (m *Manager) OnNewBlock(block *util.Block) error {
	acceptedTxs, err := m.txPool.HandleNewBlock(block)
	if err != nil {
		return err
	}

	// TODO(libp2p) broadcast all acceptedTxs
	m.updateTransactionsToRebroadcast(block)
	m.maybeRebroadcastTransactions()
	return nil
}
