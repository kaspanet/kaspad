package p2p

import (
	"github.com/daglabs/btcd/peer"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/wire"
)

// OnMemPool is invoked when a peer receives a mempool bitcoin message.
// It creates and sends an inventory message with the contents of the memory
// pool up to the maximum inventory allowed per message.  When the peer has a
// bloom filter loaded, the contents are filtered accordingly.
func (sp *Peer) OnMemPool(_ *peer.Peer, msg *wire.MsgMemPool) {
	// Only allow mempool requests if the server has bloom filtering
	// enabled.
	if sp.server.services&wire.SFNodeBloom != wire.SFNodeBloom {
		peerLog.Debugf("peer %s sent mempool request with bloom "+
			"filtering disabled -- disconnecting", sp)
		sp.Disconnect()
		return
	}

	// A decaying ban score increase is applied to prevent flooding.
	// The ban score accumulates and passes the ban threshold if a burst of
	// mempool messages comes from a peer. The score decays each minute to
	// half of its value.
	sp.addBanScore(0, 33, "mempool")

	// Generate inventory message with the available transactions in the
	// transaction memory pool.  Limit it to the max allowed inventory
	// per message.  The NewMsgInvSizeHint function automatically limits
	// the passed hint to the maximum allowed, so it's safe to pass it
	// without double checking it here.
	txMemPool := sp.server.TxMemPool
	txDescs := txMemPool.TxDescs()
	invMsg := wire.NewMsgInvSizeHint(uint(len(txDescs)))

	for _, txDesc := range txDescs {
		// Either add all transactions when there is no bloom filter,
		// or only the transactions that match the filter when there is
		// one.
		if !sp.filter.IsLoaded() || sp.filter.MatchTxAndUpdate(txDesc.Tx) {
			iv := wire.NewInvVect(wire.InvTypeTx, (*daghash.Hash)(txDesc.Tx.ID()))
			invMsg.AddInvVect(iv)
			if len(invMsg.InvList)+1 > wire.MaxInvPerMsg {
				break
			}
		}
	}

	// Send the inventory message if there is anything to send.
	if len(invMsg.InvList) > 0 {
		sp.QueueMessage(invMsg, nil)
	}
}
