package protocol

import (
	"github.com/kaspanet/kaspad/blockdag"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"sync/atomic"
	"time"
)

// StartIBDIfRequired selects a peer and starts IBD against it
// if required
func (m *Manager) StartIBDIfRequired() {
	m.startIBDMutex.Lock()
	defer m.startIBDMutex.Unlock()

	if m.IsInIBD() {
		return
	}

	peer := m.selectPeerForIBD(m.dag)
	if peer == nil {
		m.requestSelectedTipsIfRequired()
		return
	}

	atomic.StoreUint32(&m.isInIBD, 1)
	peer.StartIBD()
}

// IsInIBD is true if IBD is currently running
func (m *Manager) IsInIBD() bool {
	return atomic.LoadUint32(&m.isInIBD) != 0
}

// selectPeerForIBD returns the first peer whose selected tip
// hash is not in our DAG
func (m *Manager) selectPeerForIBD(dag *blockdag.BlockDAG) *peerpkg.Peer {
	for _, peer := range m.readyPeers {
		peerSelectedTipHash := peer.SelectedTipHash()
		if !dag.IsInDAG(peerSelectedTipHash) {
			return peer
		}
	}
	return nil
}

func (m *Manager) requestSelectedTipsIfRequired() {
	if m.isDAGTimeCurrent() {
		return
	}
	m.requestSelectedTips()
}

func (m *Manager) isDAGTimeCurrent() bool {
	const minDurationToRequestSelectedTips = time.Minute
	return m.dag.Now().Sub(m.dag.SelectedTipHeader().Timestamp) > minDurationToRequestSelectedTips
}

func (m *Manager) requestSelectedTips() {
	for _, peer := range m.readyPeers {
		peer.RequestSelectedTipIfRequired()
	}
}

func (m *Manager) FinishIBD() {
	atomic.StoreUint32(&m.isInIBD, 0)

	m.StartIBDIfRequired()
}
