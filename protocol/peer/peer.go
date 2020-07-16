package peer

import (
	"github.com/kaspanet/kaspad/netadapter/id"
	"github.com/kaspanet/kaspad/util/daghash"
	mathUtil "github.com/kaspanet/kaspad/util/math"
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

	id                    *id.ID
	userAgent             string
	services              wire.ServiceFlag
	advertisedProtocolVer uint32 // protocol version advertised by remote
	protocolVersion       uint32 // negotiated protocol version
	disableRelayTx        bool
	subnetworkID          *subnetworkid.SubnetworkID

	pingLock         sync.RWMutex
	lastPingNonce    uint64        // The nonce of the last ping we sent
	lastPingTime     time.Time     // Time we sent last ping
	lastPingDuration time.Duration // Time for last ping to return

	ibdStartChan chan struct{}
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

// ID returns the peer ID.
func (p *Peer) ID() (*id.ID, error) {
	if atomic.LoadUint32(&p.ready) == 0 {
		return nil, errors.New("peer is not ready yet")
	}
	return p.id, nil
}

// MarkAsReady marks the peer as ready.
func (p *Peer) MarkAsReady() error {
	if atomic.AddUint32(&p.ready, 1) != 1 {
		return errors.New("peer is already ready")
	}
	return nil
}

// UpdateFieldsFromMsgVersion updates the peer with the data from the version message.
func (p *Peer) UpdateFieldsFromMsgVersion(msg *wire.MsgVersion) {
	// Negotiate the protocol version.
	p.advertisedProtocolVer = msg.ProtocolVersion
	p.protocolVersion = mathUtil.MinUint32(p.protocolVersion, p.advertisedProtocolVer)
	log.Debugf("Negotiated protocol version %d for peer %s",
		p.protocolVersion, p)

	// Set the peer's ID.
	p.id = msg.ID

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
	p.lastPingDuration = time.Since(p.lastPingTime)
}

func (p *Peer) String() string {
	//TODO(libp2p)
	panic("unimplemented")
}

var (
	readyPeers      = make(map[*id.ID]*Peer, 0)
	readyPeersMutex sync.RWMutex
)

// ErrPeerWithSameIDExists signifies that a peer with the same ID already exist.
var ErrPeerWithSameIDExists = errors.New("ready with the same ID already exists")

// AddToReadyPeers marks this peer as ready and adds it to the ready peers list.
func AddToReadyPeers(peer *Peer) error {
	readyPeersMutex.RLock()
	defer readyPeersMutex.RUnlock()

	if _, ok := readyPeers[peer.id]; ok {
		return errors.Wrapf(ErrPeerWithSameIDExists, "peer with ID %s already exists", peer.id)
	}

	err := peer.MarkAsReady()
	if err != nil {
		return err
	}

	readyPeers[peer.id] = peer
	return nil
}

// GetReadyPeerIDs returns the peer IDs of all the ready peers.
func GetReadyPeerIDs() []*id.ID {
	readyPeersMutex.RLock()
	defer readyPeersMutex.RUnlock()
	peerIDs := make([]*id.ID, len(readyPeers))
	i := 0
	for peerID := range readyPeers {
		peerIDs[i] = peerID
		i++
	}
	return peerIDs
}

// IDExists returns whether there's a peer with the given ID.
func IDExists(peerID *id.ID) bool {
	_, ok := readyPeers[peerID]
	return ok
}

// ReadyPeers returns a copy of the currently ready peers
func ReadyPeers() []*Peer {
	peers := make([]*Peer, 0, len(readyPeers))
	for _, readyPeer := range readyPeers {
		peers = append(peers, readyPeer)
	}
	return peers
}

func (p *Peer) StartIBD() {
	p.ibdStartChan <- struct{}{}
}

func (p *Peer) WaitForIBDStart() {
	<-p.ibdStartChan
}
