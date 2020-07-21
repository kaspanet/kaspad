package flowcontext

import (
	"github.com/kaspanet/kaspad/netadapter"
	"github.com/kaspanet/kaspad/netadapter/id"
	"github.com/kaspanet/kaspad/protocol/common"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

// NetAdapter returns the net adapter that is associated to the flow context.
func (f *FlowContext) NetAdapter() *netadapter.NetAdapter {
	return f.netAdapter
}

// AddToReadyPeers marks this peer as ready and adds it to the ready peers list.
func (f *FlowContext) AddToReadyPeers(peer *peerpkg.Peer) error {
	f.readyPeersMutex.RLock()
	defer f.readyPeersMutex.RUnlock()

	if _, ok := f.readyPeers[peer.ID()]; ok {
		return errors.Wrapf(common.ErrPeerWithSameIDExists, "peer with ID %s already exists", peer.ID())
	}

	f.readyPeers[peer.ID()] = peer
	return nil
}

// readyPeerIDs returns the peer IDs of all the ready peers.
func (f *FlowContext) readyPeerIDs() []*id.ID {
	f.readyPeersMutex.RLock()
	defer f.readyPeersMutex.RUnlock()
	peerIDs := make([]*id.ID, len(f.readyPeers))
	i := 0
	for peerID := range f.readyPeers {
		peerIDs[i] = peerID
		i++
	}
	return peerIDs
}

// Broadcast broadcast the given message to all the ready peers.
func (f *FlowContext) Broadcast(message wire.Message) error {
	return f.netAdapter.Broadcast(f.readyPeerIDs(), message)
}
