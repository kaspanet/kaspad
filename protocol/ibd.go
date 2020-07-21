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

	peer := selectPeerForIBD(m.dag)
	if peer == nil {
		requestSelectedTipsIfRequired(m.dag)
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
func selectPeerForIBD(dag *blockdag.BlockDAG) *peerpkg.Peer {
	for _, peer := range peerpkg.ReadyPeers() {
		peerSelectedTipHash := peer.SelectedTipHash()
		if !dag.IsInDAG(peerSelectedTipHash) {
			return peer
		}
	}
	return nil
}

func requestSelectedTipsIfRequired(dag *blockdag.BlockDAG) {
	if isDAGTimeCurrent(dag) {
		return
	}
	requestSelectedTips()
}

func isDAGTimeCurrent(dag *blockdag.BlockDAG) bool {
	const minDurationToRequestSelectedTips = time.Minute
	return dag.Now().Sub(dag.SelectedTipHeader().Timestamp) > minDurationToRequestSelectedTips
}

func requestSelectedTips() {
	for _, peer := range peerpkg.ReadyPeers() {
		peer.RequestSelectedTipIfRequired()
	}
}

func (m *Manager) FinishIBD() {
	atomic.StoreUint32(&m.isInIBD, 0)

	m.StartIBDIfRequired()
}
