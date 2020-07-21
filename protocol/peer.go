package protocol

import (
	"github.com/kaspanet/kaspad/netadapter/id"
	"github.com/kaspanet/kaspad/protocol/common"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

// AddToReadyPeers marks this peer as ready and adds it to the ready peers list.
func (m *Manager) AddToReadyPeers(peer *peerpkg.Peer) error {
	m.readyPeersMutex.RLock()
	defer m.readyPeersMutex.RUnlock()

	if _, ok := m.readyPeers[peer.ID()]; ok {
		return errors.Wrapf(common.ErrPeerWithSameIDExists, "peer with ID %s already exists", peer.ID())
	}

	m.readyPeers[peer.ID()] = peer
	return nil
}

// readyPeerIDs returns the peer IDs of all the ready peers.
func (m *Manager) readyPeerIDs() []*id.ID {
	m.readyPeersMutex.RLock()
	defer m.readyPeersMutex.RUnlock()
	peerIDs := make([]*id.ID, len(m.readyPeers))
	i := 0
	for peerID := range m.readyPeers {
		peerIDs[i] = peerID
		i++
	}
	return peerIDs
}

func (m *Manager) Broadcast(message wire.Message) error {
	return m.netAdapter.Broadcast(m.readyPeerIDs(), message)
}
