package p2p

import (
	"github.com/daglabs/btcd/peer"
	"github.com/daglabs/btcd/wire"
)

// OnGetBlockLocator is invoked when a peer receives a getlocator bitcoin
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
