package peer

import (
	"github.com/kaspanet/kaspad/netadapter/id"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/subnetworkid"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
	"sync"
	"sync/atomic"
	"time"
)

// Peer holds data about a peer.
type Peer struct {
	ready uint32

	selectedTipHashMtx sync.RWMutex
	selectedTipHash    *daghash.Hash

	id                    uint32
	userAgent             string
	services              wire.ServiceFlag
	advertisedProtocolVer uint32 // protocol version advertised by remote
	protocolVersion       uint32 // negotiated protocol version
	disableRelayTx        bool
	subnetworkID          *subnetworkid.SubnetworkID

	pingLock       sync.RWMutex
	lastPingNonce  uint64    // Set to nonce if we have a pending ping.
	lastPingTime   time.Time // Time we sent last ping.
	lastPingMicros int64     // Time for last ping to return.
}

// SelectedTipHash returns the selected tip of the peer.
func (p *Peer) SelectedTipHash() (*daghash.Hash, error) {
	if atomic.LoadUint32(&p.ready) == 0 {
		return nil, errors.New("peer is not ready yet")
	}
	p.selectedTipHashMtx.RLock()
	defer p.selectedTipHashMtx.RUnlock()
	return p.selectedTipHash, nil
}

// SetSelectedTipHash sets the selected tip of the peer.
func (p *Peer) SetSelectedTipHash(hash *daghash.Hash) error {
	if atomic.LoadUint32(&p.ready) == 0 {
		return errors.New("peer is not ready yet")
	}
	p.selectedTipHashMtx.Lock()
	defer p.selectedTipHashMtx.Unlock()
	p.selectedTipHash = hash
	return nil
}

// SubnetworkID returns the subnetwork the peer is associated with.
// It is nil in full nodes.
func (p *Peer) SubnetworkID() (*subnetworkid.SubnetworkID, error) {
	if atomic.LoadUint32(&p.ready) == 0 {
		return nil, errors.New("peer is not ready yet")
	}
	return p.subnetworkID, nil
}

// MarkAsReady marks the peer as ready.
func (p *Peer) MarkAsReady() error {
	if atomic.AddUint32(&p.ready, 1) != 1 {
		return errors.New("peer is already ready")
	}
	return nil
}

// UpdateFieldsFromMsgVersion updates the peer with the data from the version message.
func (p *Peer) UpdateFieldsFromMsgVersion(msg *wire.MsgVersion, peerID uint32) {
	// Negotiate the protocol version.
	p.advertisedProtocolVer = msg.ProtocolVersion
	p.protocolVersion = minUint32(p.protocolVersion, p.advertisedProtocolVer)
	log.Debugf("Negotiated protocol version %d for peer %s",
		p.protocolVersion, p)

	// Set the peer's ID.
	p.id = peerID

	// Set the supported services for the peer to what the remote peer
	// advertised.
	p.services = msg.Services

	// Set the remote peer's user agent.
	p.userAgent = msg.UserAgent

	p.disableRelayTx = msg.DisableRelayTx
	p.selectedTipHash = msg.SelectedTipHash
	p.subnetworkID = msg.SubnetworkID
}

// SetPingPending sets the ping state of the peer to 'pending'
func (p *Peer) SetPingPending(nonce uint64) {
	p.pingLock.Lock()
	defer p.pingLock.Unlock()

	p.lastPingNonce = nonce
	p.lastPingTime = time.Now()
}

// SetPingIdle sets the ping state of the peer to 'idle'
func (p *Peer) SetPingIdle() {
	p.pingLock.Lock()
	defer p.pingLock.Unlock()

	p.lastPingNonce = 0
	p.lastPingMicros = time.Since(p.lastPingTime).Nanoseconds() / 1000

}

func (p *Peer) String() string {
	//TODO(libp2p)
	panic("unimplemented")
}

// minUint32 is a helper function to return the minimum of two uint32s.
// This avoids a math import and the need to cast to floats.
func minUint32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

// GetReadyPeerIDs returns the peer IDs of all the ready peers.
func GetReadyPeerIDs() []*id.ID {
	// TODO(libp2p)
	panic("unimplemented")
}
