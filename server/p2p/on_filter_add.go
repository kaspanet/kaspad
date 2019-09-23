package p2p

import (
	"github.com/daglabs/btcd/peer"
	"github.com/daglabs/btcd/wire"
)

// OnFilterAdd is invoked when a peer receives a filteradd bitcoin
// message and is used by remote peers to add data to an already loaded bloom
// filter.  The peer will be disconnected if a filter is not loaded when this
// message is received or the server is not configured to allow bloom filters.
func (sp *Peer) OnFilterAdd(_ *peer.Peer, msg *wire.MsgFilterAdd) {
	// Disconnect and/or ban depending on the node bloom services flag and
	// negotiated protocol version.
	if !sp.enforceNodeBloomFlag(msg.Command()) {
		return
	}

	if sp.filter.IsLoaded() {
		peerLog.Debugf("%s sent a filteradd request with no filter "+
			"loaded -- disconnecting", sp)
		sp.Disconnect()
		return
	}

	sp.filter.Add(msg.Data)
}
