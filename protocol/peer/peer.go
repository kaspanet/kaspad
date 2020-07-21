package peer

import (
	"github.com/kaspanet/kaspad/netadapter"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kaspanet/kaspad/netadapter/id"
	"github.com/kaspanet/kaspad/util/daghash"
	mathUtil "github.com/kaspanet/kaspad/util/math"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/kaspanet/kaspad/util/subnetworkid"
	"github.com/kaspanet/kaspad/wire"
)

// Peer holds data about a peer.
type Peer struct {
	connection *netadapter.NetConnection

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
func New(connection *netadapter.NetConnection) *Peer {
	return &Peer{
		connection:             connection,
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
	return p.connection.String()
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
	atomic.StoreUint32(&p.isSelectedTipRequested, 0)
}

// StartIBD starts the IBD process for this peer
func (p *Peer) StartIBD() {
	p.ibdStartChan <- struct{}{}
}

// WaitForIBDStart blocks the current thread until
// IBD start is requested from this peer
func (p *Peer) WaitForIBDStart() {
	<-p.ibdStartChan
}

func (p *Peer) Address() string {
	return p.connection.Address()
}

func (p *Peer) LastPingDuration() time.Duration {
	p.pingLock.Lock()
	defer p.pingLock.Unlock()

	return p.lastPingDuration
}
