package p2p

import (
	"github.com/kaspanet/kaspad/peer"
	"github.com/kaspanet/kaspad/wire"
)

// OnFilterClear is invoked when a peer receives a filterclear kaspa
// message and is used by remote peers to clear an already loaded bloom filter.
// The peer will be disconnected if a filter is not loaded when this message is
// received  or the server is not configured to allow bloom filters.
func (sp *Peer) OnFilterClear(_ *peer.Peer, msg *wire.MsgFilterClear) {
	// Disconnect and/or ban depending on the node bloom services flag and
	// negotiated protocol version.
	if !sp.enforceNodeBloomFlag(msg.Command()) {
		return
	}

	if !sp.filter.IsLoaded() {
		sp.AddBanScoreAndPushRejectMsg(wire.CmdFilterClear, wire.RejectInvalid, nil,
			peer.BanScoreNoFilterLoaded, 0, "sent a filterclear request with no filter loaded")
		return
	}

	sp.filter.Unload()
}
