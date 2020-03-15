package p2p

import (
	"github.com/kaspanet/kaspad/addrmgr"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/peer"
	"github.com/kaspanet/kaspad/wire"
)

// OnVersion is invoked when a peer receives a version kaspa message
// and is used to negotiate the protocol version details as well as kick start
// the communications.
func (sp *Peer) OnVersion(_ *peer.Peer, msg *wire.MsgVersion) {
	// Signal the sync manager this peer is a new sync candidate.
	sp.server.SyncManager.NewPeer(sp.Peer)

	// Choose whether or not to relay transactions before a filter command
	// is received.
	sp.setDisableRelayTx(msg.DisableRelayTx)

	// Update the address manager and request known addresses from the
	// remote peer for outbound connections. This is skipped when running
	// on the simulation test network since it is only intended to connect
	// to specified peers and actively avoids advertising and connecting to
	// discovered peers.
	if !config.ActiveConfig().Simnet {
		addrManager := sp.server.addrManager

		// Outbound connections.
		if !sp.Inbound() {
			// TODO(davec): Only do this if not doing the initial block
			// download and the local address is routable.
			if !config.ActiveConfig().DisableListen /* && isCurrent? */ {
				// Get address that best matches.
				lna := addrManager.GetBestLocalAddress(sp.NA())
				if addrmgr.IsRoutable(lna) {
					// Filter addresses the peer already knows about.
					addresses := []*wire.NetAddress{lna}
					sp.pushAddrMsg(addresses, sp.SubnetworkID())
				}
			}

			// Request known addresses if the server address manager needs
			// more.
			if addrManager.NeedMoreAddresses() {
				sp.QueueMessage(wire.NewMsgGetAddr(false, sp.SubnetworkID()), nil)

				if sp.SubnetworkID() != nil {
					sp.QueueMessage(wire.NewMsgGetAddr(false, nil), nil)
				}
			}

			// Mark the address as a known good address.
			addrManager.Good(sp.NA(), msg.SubnetworkID)
		}
	}

	// Add valid peer to the server.
	sp.server.AddPeer(sp)
}
