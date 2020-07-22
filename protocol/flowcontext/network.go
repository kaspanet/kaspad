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

// AddToPeers marks this peer as ready and adds it to the ready peers list.
func (f *FlowContext) AddToPeers(peer *peerpkg.Peer) error {
	f.peersMutex.RLock()
	defer f.peersMutex.RUnlock()

	if _, ok := f.peers[peer.ID()]; ok {
		return errors.Wrapf(common.ErrPeerWithSameIDExists, "peer with ID %s already exists", peer.ID())
	}

	f.peers[peer.ID()] = peer

	if f.peerAddedCallback != nil {
		f.peerAddedCallback(peer)
	}

	return nil
}

// readyPeerIDs returns the peer IDs of all the ready peers.
func (f *FlowContext) readyPeerIDs() []*id.ID {
	f.peersMutex.RLock()
	defer f.peersMutex.RUnlock()
	peerIDs := make([]*id.ID, len(f.peers))
	i := 0
	for peerID := range f.peers {
		peerIDs[i] = peerID
		i++
	}
	return peerIDs
}

// Broadcast broadcast the given message to all the ready peers.
func (f *FlowContext) Broadcast(message wire.Message) error {
	return f.netAdapter.Broadcast(f.readyPeerIDs(), message)
}

// Peers returns the currently active peers
func (f *FlowContext) Peers() []*peerpkg.Peer {
	f.peersMutex.RLock()
	defer f.peersMutex.RUnlock()

	peers := make([]*peerpkg.Peer, len(f.peers))
	i := 0
	for _, peer := range f.peers {
		peers[i] = peer
		i++
	}
	return peers
}

type PeerAddedCallback func(*peerpkg.Peer)

func (f *FlowContext) SetPeerAddedCallback(callback PeerAddedCallback) {
	f.peerAddedCallback = callback
}
