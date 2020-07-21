package peer

import (
	"github.com/kaspanet/kaspad/netadapter/id"
	"github.com/pkg/errors"
	"sync"
)

// Peers holds a list of active peers
type Peers struct {
	readyPeers      map[*id.ID]*Peer
	readyPeersMutex sync.RWMutex
}

// NewPeers returns a new Peers
func NewPeers() *Peers {
	return &Peers{
		readyPeers: make(map[*id.ID]*Peer, 0),
	}
}

// ErrPeerWithSameIDExists signifies that a peer with the same ID already exist.
var ErrPeerWithSameIDExists = errors.New("ready with the same ID already exists")

// AddToReadyPeers marks this peer as ready and adds it to the ready peers list.
func (p *Peers) AddToReadyPeers(peer *Peer) error {
	p.readyPeersMutex.RLock()
	defer p.readyPeersMutex.RUnlock()

	if _, ok := p.readyPeers[peer.id]; ok {
		return errors.Wrapf(ErrPeerWithSameIDExists, "peer with ID %s already exists", peer.id)
	}

	p.readyPeers[peer.id] = peer
	return nil
}

// ReadyPeerIDs returns the peer IDs of all the ready peers.
func (p *Peers) ReadyPeerIDs() []*id.ID {
	p.readyPeersMutex.RLock()
	defer p.readyPeersMutex.RUnlock()
	peerIDs := make([]*id.ID, len(p.readyPeers))
	i := 0
	for peerID := range p.readyPeers {
		peerIDs[i] = peerID
		i++
	}
	return peerIDs
}

// ReadyPeers returns a copy of the currently ready peers
func (p *Peers) ReadyPeers() []*Peer {
	peers := make([]*Peer, 0, len(p.readyPeers))
	for _, readyPeer := range p.readyPeers {
		peers = append(peers, readyPeer)
	}
	return peers
}
