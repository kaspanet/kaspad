package p2p

import (
	"github.com/kaspanet/kaspad/peer"
	"github.com/kaspanet/kaspad/wire"
)

// OnGetBlockLocator is invoked when a peer receives a getlocator kaspa
// message.
func (sp *Peer) OnGetBlockLocator(_ *peer.Peer, msg *wire.MsgGetBlockLocator) {
	locator := sp.server.DAG.BlockLocatorFromHashes(msg.StartHash, msg.StopHash)

	if len(locator) == 0 {
		peerLog.Infof("Couldn't build a block locator between blocks %s and %s"+
			" that was requested from peer %s",
			sp)
		return
	}
	err := sp.PushBlockLocatorMsg(locator)
	if err != nil {
		peerLog.Errorf("Failed to send block locator message to peer %s: %s",
			sp, err)
		return
	}
}
