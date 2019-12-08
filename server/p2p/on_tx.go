package p2p

import (
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/peer"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

// OnTx is invoked when a peer receives a tx bitcoin message.  It blocks
// until the bitcoin transaction has been fully processed.  Unlock the block
// handler this does not serialize all transactions through a single thread
// transactions don't rely on the previous one in a linear fashion like blocks.
func (sp *Peer) OnTx(_ *peer.Peer, msg *wire.MsgTx) {
	if config.ActiveConfig().BlocksOnly {
		peerLog.Tracef("Ignoring tx %s from %s - blocksonly enabled",
			msg.TxID(), sp)
		return
	}

	// Add the transaction to the known inventory for the peer.
	// Convert the raw MsgTx to a util.Tx which provides some convenience
	// methods and things such as hash caching.
	tx := util.NewTx(msg)
	iv := wire.NewInvVect(wire.InvTypeTx, (*daghash.Hash)(tx.ID()))
	sp.AddKnownInventory(iv)

	// Queue the transaction up to be handled by the sync manager and
	// intentionally block further receives until the transaction is fully
	// processed and known good or bad.  This helps prevent a malicious peer
	// from queuing up a bunch of bad transactions before disconnecting (or
	// being disconnected) and wasting memory.
	sp.server.SyncManager.QueueTx(tx, sp.Peer, sp.txProcessed)
	<-sp.txProcessed
}
