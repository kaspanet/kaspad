package p2p

import (
	"github.com/kaspanet/kaspad/peer"
	"github.com/kaspanet/kaspad/wire"
)

// OnGetBlockInvs is invoked when a peer receives a getblockinvs kaspa
// message.
// It finds the blue future between msg.StartHash and msg.StopHash
// and send the invs to the requesting peer.
func (sp *Peer) OnGetBlockInvs(_ *peer.Peer, msg *wire.MsgGetBlockInvs) {
	dag := sp.server.DAG
	// We want to prevent a situation where the syncing peer needs
	// to call getblocks once again, but the block we sent him
	// won't affect his selected chain, so next time it'll try
	// to find the highest shared chain block, it'll find the
	// same one as before.
	// To prevent that we use blockdag.FinalityInterval as maxHashes.
	// This way, if one getblocks is not enough to get the peer
	// synced, we can know for sure that its selected chain will
	// change, so we'll have higher shared chain block.
	hashList, err := dag.GetBlueBlocksHashesBetween(msg.StartHash, msg.StopHash,
		wire.MaxInvPerMsg)
	if err != nil {
		peerLog.Warnf("Error getting blue blocks between %s and %s: %s", msg.StartHash, msg.StopHash, err)
		sp.Disconnect()
		return
	}

	// Generate inventory message.
	invMsg := wire.NewMsgInv()
	for i := range hashList {
		iv := wire.NewInvVect(wire.InvTypeSyncBlock, hashList[i])
		invMsg.AddInvVect(iv)
	}

	// Send the inventory message if there is anything to send.
	if len(invMsg.InvList) > 0 {
		sp.QueueMessage(invMsg, nil)
	}
}
