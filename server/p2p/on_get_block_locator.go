package p2p

import (
	"fmt"
	"github.com/kaspanet/kaspad/peer"
	"github.com/kaspanet/kaspad/wire"
)

// OnGetBlockLocator is invoked when a peer receives a getlocator kaspa
// message.
func (sp *Peer) OnGetBlockLocator(_ *peer.Peer, msg *wire.MsgGetBlockLocator) {
	locator, err := sp.server.DAG.BlockLocatorFromHashes(msg.HighHash, msg.LowHash)
	if err != nil || len(locator) == 0 {
		if err != nil {
			peerLog.Warnf("Couldn't build a block locator between blocks "+
				"%s and %s that was requested from peer %s: %s", msg.HighHash, msg.LowHash, sp, err)
		}
		sp.AddBanScoreAndPushRejectMsg(msg.Command(), wire.RejectInvalid, nil, 100, 0, fmt.Sprintf("couldn't build a block locator between blocks %s and %s", msg.HighHash, msg.LowHash))
		return
	}

	err = sp.PushBlockLocatorMsg(locator)
	if err != nil {
		peerLog.Errorf("Failed to send block locator message to peer %s: %s",
			sp, err)
		return
	}
}
