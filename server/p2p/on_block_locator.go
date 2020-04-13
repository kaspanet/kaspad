package p2p

import (
	"github.com/kaspanet/kaspad/peer"
	"github.com/kaspanet/kaspad/wire"
)

// OnBlockLocator is invoked when a peer receives a locator kaspa
// message.
func (sp *Peer) OnBlockLocator(_ *peer.Peer, msg *wire.MsgBlockLocator) {
	sp.SetWasBlockLocatorRequested(false)
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
	highHash := msg.BlockLocatorHashes[0]
	if dag.IsInDAG(highHash) {
		if dag.IsKnownFinalizedBlock(highHash) {
			peerLog.Debugf("Cannot sync with peer %s because the highest"+
				" shared chain block (%s) is below the finality point", sp, highHash)
			sp.server.SyncManager.RemoveFromSyncCandidates(sp.Peer)
			return
		}

		// We send the highHash as the GetBlockInvsMsg's lowHash here.
		// This is not a mistake. The invs we desire start from the highest
		// hash that we know of and end at the highest hash that the peer
		// knows of.
		err := sp.Peer.PushGetBlockInvsMsg(highHash, sp.Peer.SelectedTipHash())
		if err != nil {
			peerLog.Errorf("Failed pushing get blocks message for peer %s: %s",
				sp, err)
			return
		}
		return
	}
	highHash, lowHash := dag.FindNextLocatorBoundaries(msg.BlockLocatorHashes)
	if highHash == nil {
		panic("Couldn't find any unknown hashes in the block locator.")
	}
	sp.PushGetBlockLocatorMsg(highHash, lowHash)
}
