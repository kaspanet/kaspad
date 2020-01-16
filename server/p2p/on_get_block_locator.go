package p2p

import (
	"fmt"
	"github.com/kaspanet/kaspad/peer"
	"github.com/kaspanet/kaspad/wire"
)

// OnGetBlockLocator is invoked when a peer receives a getlocator kaspa
// message.
func (sp *Peer) OnGetBlockLocator(_ *peer.Peer, msg *wire.MsgGetBlockLocator) {
	locator, err := sp.server.DAG.BlockLocatorFromHashes(msg.StartHash, msg.StopHash)
	if err != nil || len(locator) == 0 {
		warning := fmt.Sprintf("Couldn't build a block locator between blocks "+
			"%s and %s that was requested from peer %s", msg.StartHash, msg.StopHash, sp)
		if err != nil {
			warning = fmt.Sprintf("%s: %s", warning, err)
		}
		peerLog.Warnf(warning)
		sp.Disconnect()
		return
	}

	err = sp.PushBlockLocatorMsg(locator)
	if err != nil {
		peerLog.Errorf("Failed to send block locator message to peer %s: %s",
			sp, err)
		return
	}
}
