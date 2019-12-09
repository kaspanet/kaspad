package p2p

import (
	"github.com/kaspanet/kaspad/peer"
	"github.com/kaspanet/kaspad/wire"
	"github.com/prometheus/common/log"
)

// OnGetHeaders is invoked when a peer receives a getheaders bitcoin
// message.
func (sp *Peer) OnGetHeaders(_ *peer.Peer, msg *wire.MsgGetHeaders) {
	// Ignore getheaders requests if not in sync.
	if !sp.server.SyncManager.IsCurrent() {
		return
	}

	// Find the most recent known block in the best chain based on the block
	// locator and fetch all of the headers after it until either
	// wire.MaxBlockHeadersPerMsg have been fetched or the provided stop
	// hash is encountered.
	//
	// Use the block after the genesis block if no other blocks in the
	// provided locator are known. This does mean the client will start
	// over with the genesis block if unknown block locators are provided.
	dag := sp.server.DAG
	headers, err := dag.GetBlueBlocksHeadersBetween(msg.StartHash, msg.StopHash)
	if err != nil {
		log.Warnf("Error getting blue blocks headers between %s and %s: %s", msg.StartHash, msg.StopHash, err)
		sp.Disconnect()
		return
	}

	// Send found headers to the requesting peer.
	blockHeaders := make([]*wire.BlockHeader, len(headers))
	for i := range headers {
		blockHeaders[i] = headers[i]
	}
	sp.QueueMessage(&wire.MsgHeaders{Headers: blockHeaders}, nil)
}
