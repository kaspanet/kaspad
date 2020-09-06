package mining

// This file functions are not considered safe for regular use, and should be used for test purposes only.

import (
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/mstime"
)

// fakeTxSource is a simple implementation of TxSource interface
type fakeTxSource struct {
	txDescs []*TxDesc
}

func (txs *fakeTxSource) LastUpdated() mstime.Time {
	return mstime.UnixMilliseconds(0)
}

func (txs *fakeTxSource) MiningDescs() []*TxDesc {
	return txs.txDescs
}

func (txs *fakeTxSource) HaveTransaction(txID *daghash.TxID) bool {
	for _, desc := range txs.txDescs {
		if *desc.Tx.ID() == *txID {
			return true
		}
	}
	return false
}
