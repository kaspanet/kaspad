package flowcontext

import (
	"sync/atomic"
	"time"

	"github.com/kaspanet/kaspad/util/mstime"

	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
)

// StartIBDIfRequired selects a peer and starts IBD against it
// if required
func (f *FlowContext) StartIBDIfRequired() error {
	f.startIBDMutex.Lock()
	defer f.startIBDMutex.Unlock()

	if f.IsInIBD() {
		return nil
	}

	peer, err := f.selectPeerForIBD()
	if err != nil {
		return err
	}
	if peer == nil {
		spawn("StartIBDIfRequired-requestSelectedTipsIfRequired", f.requestSelectedTipsIfRequired)
		return nil
	}

	atomic.StoreUint32(&f.isInIBD, 1)
	f.ibdPeer = peer
	spawn("StartIBDIfRequired-peer.StartIBD", peer.StartIBD)

	return nil
}

// IsInIBD is true if IBD is currently running
func (f *FlowContext) IsInIBD() bool {
	return atomic.LoadUint32(&f.isInIBD) != 0
}

// selectPeerForIBD returns the first peer whose selected tip
// hash is not in our DAG
func (f *FlowContext) selectPeerForIBD() (*peerpkg.Peer, error) {
	f.peersMutex.RLock()
	defer f.peersMutex.RUnlock()

	for _, peer := range f.peers {
		peerSelectedTipHash := peer.SelectedTipHash()
		blockInfo, err := f.domain.Consensus().GetBlockInfo(peerSelectedTipHash)
		if err != nil {
			return nil, err
		}
		if !blockInfo.Exists {
			return peer, nil
		}
	}
	return nil, nil
}

func (f *FlowContext) requestSelectedTipsIfRequired() {
	dagTimeCurrent, err := f.isDAGTimeCurrent()
	if err != nil {
		panic(err)
	}
	if dagTimeCurrent {
		return
	}
	f.requestSelectedTips()
}

func (f *FlowContext) isDAGTimeCurrent() (bool, error) {
	const minDurationToRequestSelectedTips = time.Minute
	virtualSelectedParent, err := f.domain.Consensus().GetVirtualSelectedParent()
	if err != nil {
		return false, err
	}
	virtualSelectedParentTime := mstime.UnixMilliseconds(virtualSelectedParent.Header.TimeInMilliseconds)
	return mstime.Now().Sub(virtualSelectedParentTime) > minDurationToRequestSelectedTips, nil
}

func (f *FlowContext) requestSelectedTips() {
	f.peersMutex.RLock()
	defer f.peersMutex.RUnlock()

	for _, peer := range f.peers {
		peer.RequestSelectedTipIfRequired()
	}
}

// FinishIBD finishes the current IBD flow and starts a new one if required.
func (f *FlowContext) FinishIBD() error {
	f.ibdPeer = nil

	atomic.StoreUint32(&f.isInIBD, 0)

	return f.StartIBDIfRequired()
}

// IBDPeer returns the currently active IBD peer.
// Returns nil if we aren't currently in IBD
func (f *FlowContext) IBDPeer() *peerpkg.Peer {
	if !f.IsInIBD() {
		return nil
	}
	return f.ibdPeer
}
