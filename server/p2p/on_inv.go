package p2p

import (
	"github.com/daglabs/kaspad/config"
	"github.com/daglabs/kaspad/peer"
	"github.com/daglabs/kaspad/wire"
)

// OnInv is invoked when a peer receives an inv bitcoin message and is
// used to examine the inventory being advertised by the remote peer and react
// accordingly.  We pass the message down to blockmanager which will call
// QueueMessage with any appropriate responses.
func (sp *Peer) OnInv(_ *peer.Peer, msg *wire.MsgInv) {
	if !config.ActiveConfig().BlocksOnly {
		if len(msg.InvList) > 0 {
			sp.server.SyncManager.QueueInv(msg, sp.Peer)
		}
		return
	}

	newInv := wire.NewMsgInvSizeHint(uint(len(msg.InvList)))
	for _, invVect := range msg.InvList {
		if invVect.Type == wire.InvTypeTx {
			peerLog.Tracef("Ignoring tx %s in inv from %s -- "+
				"blocksonly enabled", invVect.Hash, sp)
			peerLog.Infof("Peer %s is announcing "+
				"transactions -- disconnecting", sp)
			sp.Disconnect()
			return
		}
		err := newInv.AddInvVect(invVect)
		if err != nil {
			peerLog.Errorf("Failed to add inventory vector: %s", err)
			break
		}
	}

	if len(newInv.InvList) > 0 {
		sp.server.SyncManager.QueueInv(newInv, sp.Peer)
	}
}
