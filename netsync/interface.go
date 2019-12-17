// Copyright (c) 2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package netsync

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/mempool"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/wire"
)

// PeerNotifier exposes methods to notify peers of status changes to
// transactions, blocks, etc. Currently server (in the main package) implements
// this interface.
type PeerNotifier interface {
	AnnounceNewTransactions(newTxs []*mempool.TxDesc)

	RelayInventory(invVect *wire.InvVect, data interface{})

	TransactionConfirmed(tx *util.Tx)
}

// Config is a configuration struct used to initialize a new SyncManager.
type Config struct {
	PeerNotifier PeerNotifier
	DAG          *blockdag.BlockDAG
	TxMemPool    *mempool.TxPool
	DAGParams    *dagconfig.Params
	MaxPeers     int
}
