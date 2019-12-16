package p2p

import (
	"github.com/kaspanet/kaspad/peer"
	"github.com/kaspanet/kaspad/wire"
)

// OnBlockLocator is invoked when a peer receives a locator kaspa
// message.
func (sp *Peer) OnBlockLocator(_ *peer.Peer, msg *wire.MsgBlockLocator) {
	// Find the highest known shared block between the peers, and asks
	// the block and its future from the peer. If the block is not
	// found, create a lower resolution block locator and send it to
	// the peer in order to find it in the next iteration.
	dag := sp.server.DAG
	if len(msg.BlockLocatorHashes) == 0 {
		peerLog.Warnf("Got empty block locator from peer %s",
			sp)
		return
	}
	// If the first hash of the block locator is known, it means we found
	// the highest shared block.
	firstHash := msg.BlockLocatorHashes[0]
	if dag.BlockExists(firstHash) {
		if dag.IsKnownFinalizedBlock(firstHash) {
			peerLog.Debugf("Cannot sync with peer %s because the highest"+
				" shared chain block (%s) is below the finality point", sp, firstHash)
			sp.server.SyncManager.RemoveFromSyncCandidates(sp.Peer)
			return
		}
		err := sp.Peer.PushGetBlockInvsMsg(firstHash, sp.Peer.SelectedTip())
		if err != nil {
			peerLog.Errorf("Failed pushing get blocks message for peer %s: %s",
				sp, err)
			return
		}
		return
	}
	startHash, stopHash := dag.FindNextLocatorBoundaries(msg.BlockLocatorHashes)
	if startHash == nil {
		panic("Couldn't find any unknown hashes in the block locator.")
	}
	sp.PushGetBlockLocatorMsg(startHash, stopHash)
}
