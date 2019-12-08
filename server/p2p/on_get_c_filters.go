package p2p

import (
	"github.com/kaspanet/kaspad/peer"
	"github.com/kaspanet/kaspad/wire"
)

// OnGetCFilters is invoked when a peer receives a getcfilters bitcoin message.
func (sp *Peer) OnGetCFilters(_ *peer.Peer, msg *wire.MsgGetCFilters) {
	// Ignore getcfilters requests if not in sync.
	if !sp.server.SyncManager.IsCurrent() {
		return
	}

	hashes, err := sp.server.DAG.HeightToHashRange(msg.StartHeight,
		msg.StopHash, wire.MaxGetCFiltersReqRange)
	if err != nil {
		peerLog.Debugf("Invalid getcfilters request: %s", err)
		return
	}

	filters, err := sp.server.CfIndex.FiltersByBlockHashes(hashes,
		msg.FilterType)
	if err != nil {
		peerLog.Errorf("Error retrieving cfilters: %s", err)
		return
	}

	for i, filterBytes := range filters {
		if len(filterBytes) == 0 {
			peerLog.Warnf("Could not obtain cfilter for %s", hashes[i])
			return
		}
		filterMsg := wire.NewMsgCFilter(msg.FilterType, hashes[i], filterBytes)
		sp.QueueMessage(filterMsg, nil)
	}
}
