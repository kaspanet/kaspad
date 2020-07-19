package peer

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/kaspanet/kaspad/netadapter/id"
	"github.com/kaspanet/kaspad/util/daghash"
	mathUtil "github.com/kaspanet/kaspad/util/math"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/kaspanet/kaspad/util/subnetworkid"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

// Peer holds data about a peer.
type Peer struct {
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

	isSelectedTipRequested uint32
	selectedTipRequestChan chan struct{}
	lastSelectedTipRequest mstime.Time

	ibdStartChan chan struct{}
}

// New returns a new Peer
func New() *Peer {
	return &Peer{
		selectedTipRequestChan: make(chan struct{}),
		ibdStartChan:           make(chan struct{}),
	}
}

// SelectedTipHash returns the selected tip of the peer.
func (p *Peer) SelectedTipHash() *daghash.Hash {
	p.selectedTipHashMtx.RLock()
	defer p.selectedTipHashMtx.RUnlock()
	return p.selectedTipHash
}

// SetSelectedTipHash sets the selected tip of the peer.
func (p *Peer) SetSelectedTipHash(hash *daghash.Hash) {
	p.selectedTipHashMtx.Lock()
	defer p.selectedTipHashMtx.Unlock()
	p.selectedTipHash = hash
}

// SubnetworkID returns the subnetwork the peer is associated with.
// It is nil in full nodes.
func (p *Peer) SubnetworkID() *subnetworkid.SubnetworkID {
	return p.subnetworkID
}

// ID returns the peer ID.
func (p *Peer) ID() *id.ID {
	return p.id
}

// UpdateFieldsFromMsgVersion updates the peer with the data from the version message.
func (p *Peer) UpdateFieldsFromMsgVersion(msg *wire.MsgVersion) {
	// Negotiate the protocol version.
	p.advertisedProtocolVer = msg.ProtocolVersion
	p.protocolVersion = mathUtil.MinUint32(p.protocolVersion, p.advertisedProtocolVer)
	log.Debugf("Negotiated protocol version %d for peer %s",
		p.protocolVersion, p.id)

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

// ReadyPeers returns a copy of the currently ready peers
func ReadyPeers() []*Peer {
	peers := make([]*Peer, 0, len(readyPeers))
	for _, readyPeer := range readyPeers {
		peers = append(peers, readyPeer)
	}
	return peers
}

// RequestSelectedTipIfRequired notifies the peer that requesting
// a selected tip is required. This triggers the selected tip
// request flow.
func (p *Peer) RequestSelectedTipIfRequired() {
	if atomic.SwapUint32(&p.isSelectedTipRequested, 1) != 0 {
		return
	}

	const minGetSelectedTipInterval = time.Minute
	if mstime.Since(p.lastSelectedTipRequest) < minGetSelectedTipInterval {
		return
	}

	p.lastSelectedTipRequest = mstime.Now()
	p.selectedTipRequestChan <- struct{}{}
}

// WaitForSelectedTipRequests blocks the current thread until
// a selected tip is requested from this peer
func (p *Peer) WaitForSelectedTipRequests() {
	<-p.selectedTipRequestChan
}

// FinishRequestingSelectedTip finishes requesting the selected
// tip from this peer
func (p *Peer) FinishRequestingSelectedTip() {
	atomic.SwapUint32(&p.isSelectedTipRequested, 0)
}

// StartIBD notifies the peer that starting its IBD flow is required.
// Note that the IBD flow is expected to wait using WaitForIBDStart.
func (p *Peer) StartIBD() {
	p.ibdStartChan <- struct{}{}
}

// WaitForIBDStart blocks the current thread until
// IBD start is requested from this peer
func (p *Peer) WaitForIBDStart() {
	<-p.ibdStartChan
}
