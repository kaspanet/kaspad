package peer

import (
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/subnetworkid"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
	"sync"
	"sync/atomic"
)

type Peer struct {
	ready uint32

	selectedTipHashMtx sync.RWMutex
	selectedTipHash    *daghash.Hash

	id                 uint32
	userAgent          string
	services           wire.ServiceFlag
	advertisedProtoVer uint32 // protocol version advertised by remote
	protocolVersion    uint32 // negotiated protocol version
	disableRelayTx     bool
	subnetworkID       *subnetworkid.SubnetworkID
}

func (p *Peer) SelectedTipHash() (*daghash.Hash, error) {
	if atomic.LoadUint32(&p.ready) == 0 {
		return nil, errors.New("peer is not ready yet")
	}
	p.selectedTipHashMtx.RLock()
	defer p.selectedTipHashMtx.RUnlock()
	return p.selectedTipHash, nil
}

func (p *Peer) SubnetworkID() (*subnetworkid.SubnetworkID, error) {
	if atomic.LoadUint32(&p.ready) == 0 {
		return nil, errors.New("peer is not ready yet")
	}
	return p.subnetworkID, nil
}

func (p *Peer) SetSelectedTipHash(hash *daghash.Hash) error {
	if atomic.LoadUint32(&p.ready) == 0 {
		return errors.New("peer is not ready yet")
	}
	p.selectedTipHashMtx.Lock()
	defer p.selectedTipHashMtx.Unlock()
	p.selectedTipHash = hash
	return nil
}

func (p *Peer) MarkAsReady() error {
	if atomic.AddUint32(&p.ready, 1) != 1 {
		return errors.New("peer is already ready")
	}
	return nil
}

func (p *Peer) UpdateFlagsFromVersionMsg(msg *wire.MsgVersion, peerID uint32) {
	// Negotiate the protocol version.
	p.advertisedProtoVer = msg.ProtocolVersion
	p.protocolVersion = minUint32(p.protocolVersion, p.advertisedProtoVer)
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

// minUint32 is a helper function to return the minimum of two uint32s.
// This avoids a math import and the need to cast to floats.
func minUint32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}
