package p2p

import (
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/peer"
	"github.com/kaspanet/kaspad/wire"
)

// OnGetAddr is invoked when a peer receives a getaddr kaspa message
// and is used to provide the peer with known addresses from the address
// manager.
func (sp *Peer) OnGetAddr(_ *peer.Peer, msg *wire.MsgGetAddr) {
	// Don't return any addresses when running on the simulation test
	// network. This helps prevent the network from becoming another
	// public test network since it will not be able to learn about other
	// peers that have not specifically been provided.
	if config.ActiveConfig().Simnet {
		return
	}

	// Do not accept getaddr requests from outbound peers. This reduces
	// fingerprinting attacks.
	if !sp.Inbound() {
		peerLog.Debugf("Ignoring getaddr request from outbound peer ",
			"%s", sp)
		return
	}

	// Only allow one getaddr request per connection to discourage
	// address stamping of inv announcements.
	if sp.sentAddrs {
		peerLog.Debugf("Ignoring repeated getaddr request from peer ",
			"%s", sp)
		return
	}
	sp.sentAddrs = true

	// Get the current known addresses from the address manager.
	addrCache := sp.server.addrManager.AddressCache(msg.IncludeAllSubnetworks, msg.SubnetworkID)

	// Push the addresses.
	sp.pushAddrMsg(addrCache, sp.SubnetworkID())
}
