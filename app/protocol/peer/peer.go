package peer

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/id"
	mathUtil "github.com/kaspanet/kaspad/util/math"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/kaspanet/kaspad/util/subnetworkid"
)

// Peer holds data about a peer.
type Peer struct {
	connection *netadapter.NetConnection

	selectedTipHashMtx sync.RWMutex
	selectedTipHash    *externalapi.DomainHash

	userAgent                string
	services                 appmessage.ServiceFlag
	advertisedProtocolVerion uint32 // protocol version advertised by remote
	protocolVersion          uint32 // negotiated protocol version
	disableRelayTx           bool
	subnetworkID             *subnetworkid.SubnetworkID

	timeOffset        time.Duration
	connectionStarted time.Time

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
		connectionStarted:      time.Now(),
	}
}

// Connection returns the NetConnection associated with this peer
func (p *Peer) Connection() *netadapter.NetConnection {
	return p.connection
}

// SelectedTipHash returns the selected tip of the peer.
func (p *Peer) SelectedTipHash() *externalapi.DomainHash {
	p.selectedTipHashMtx.RLock()
	defer p.selectedTipHashMtx.RUnlock()
	return p.selectedTipHash
}

// SetSelectedTipHash sets the selected tip of the peer.
func (p *Peer) SetSelectedTipHash(hash *externalapi.DomainHash) {
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
	return p.connection.ID()
}

// TimeOffset returns the peer's time offset.
func (p *Peer) TimeOffset() time.Duration {
	return p.timeOffset
}

// UserAgent returns the peer's user agent.
func (p *Peer) UserAgent() string {
	return p.userAgent
}

// AdvertisedProtocolVersion returns the peer's advertised protocol version.
func (p *Peer) AdvertisedProtocolVersion() uint32 {
	return p.advertisedProtocolVerion
}

// TimeConnected returns the time since the connection to this been has been started.
func (p *Peer) TimeConnected() time.Duration {
	return time.Since(p.connectionStarted)
}

// IsOutbound returns whether the peer is an outbound connection.
func (p *Peer) IsOutbound() bool {
	return p.connection.IsOutbound()
}

// UpdateFieldsFromMsgVersion updates the peer with the data from the version message.
func (p *Peer) UpdateFieldsFromMsgVersion(msg *appmessage.MsgVersion) {
	// Negotiate the protocol version.
	p.advertisedProtocolVerion = msg.ProtocolVersion
	p.protocolVersion = mathUtil.MinUint32(p.protocolVersion, p.advertisedProtocolVerion)
	log.Debugf("Negotiated protocol version %d for peer %s",
		p.protocolVersion, p.ID())

	// Set the supported services for the peer to what the remote peer
	// advertised.
	p.services = msg.Services

	// Set the remote peer's user agent.
	p.userAgent = msg.UserAgent

	p.disableRelayTx = msg.DisableRelayTx
	p.selectedTipHash = msg.SelectedTipHash
	p.subnetworkID = msg.SubnetworkID

	p.timeOffset = mstime.Since(msg.Timestamp)
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

// Address returns the address associated with this connection
func (p *Peer) Address() string {
	return p.connection.Address()
}

// LastPingDuration returns the duration of the last ping to
// this peer
func (p *Peer) LastPingDuration() time.Duration {
	p.pingLock.Lock()
	defer p.pingLock.Unlock()

	return p.lastPingDuration
}
