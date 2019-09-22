package p2p

import (
	"github.com/daglabs/btcd/peer"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/wire"
)

// OnGetCFHeaders is invoked when a peer receives a getcfheader bitcoin message.
func (sp *Peer) OnGetCFHeaders(_ *peer.Peer, msg *wire.MsgGetCFHeaders) {
	// Ignore getcfilterheader requests if not in sync.
	if !sp.server.SyncManager.IsCurrent() {
		return
	}

	startHeight := msg.StartHeight
	maxResults := wire.MaxCFHeadersPerMsg

	// If StartHeight is positive, fetch the predecessor block hash so we can
	// populate the PrevFilterHeader field.
	if msg.StartHeight > 0 {
		startHeight--
		maxResults++
	}

	// Fetch the hashes from the block index.
	hashList, err := sp.server.DAG.HeightToHashRange(startHeight,
		msg.StopHash, maxResults)
	if err != nil {
		peerLog.Debugf("Invalid getcfheaders request: %s", err)
	}

	// This is possible if StartHeight is one greater that the height of
	// StopHash, and we pull a valid range of hashes including the previous
	// filter header.
	if len(hashList) == 0 || (msg.StartHeight > 0 && len(hashList) == 1) {
		peerLog.Debug("No results for getcfheaders request")
		return
	}

	// Fetch the raw filter hash bytes from the database for all blocks.
	filterHashes, err := sp.server.CfIndex.FilterHashesByBlockHashes(hashList,
		msg.FilterType)
	if err != nil {
		peerLog.Errorf("Error retrieving cfilter hashes: %s", err)
		return
	}

	// Generate cfheaders message and send it.
	headersMsg := wire.NewMsgCFHeaders()

	// Populate the PrevFilterHeader field.
	if msg.StartHeight > 0 {
		parentHash := hashList[0]

		// Fetch the raw committed filter header bytes from the
		// database.
		headerBytes, err := sp.server.CfIndex.FilterHeaderByBlockHash(
			parentHash, msg.FilterType)
		if err != nil {
			peerLog.Errorf("Error retrieving CF header: %s", err)
			return
		}
		if len(headerBytes) == 0 {
			peerLog.Warnf("Could not obtain CF header for %s", parentHash)
			return
		}

		// Deserialize the hash into PrevFilterHeader.
		err = headersMsg.PrevFilterHeader.SetBytes(headerBytes)
		if err != nil {
			peerLog.Warnf("Committed filter header deserialize "+
				"failed: %s", err)
			return
		}

		hashList = hashList[1:]
		filterHashes = filterHashes[1:]
	}

	// Populate HeaderHashes.
	for i, hashBytes := range filterHashes {
		if len(hashBytes) == 0 {
			peerLog.Warnf("Could not obtain CF hash for %s", hashList[i])
			return
		}

		// Deserialize the hash.
		filterHash, err := daghash.NewHash(hashBytes)
		if err != nil {
			peerLog.Warnf("Committed filter hash deserialize "+
				"failed: %s", err)
			return
		}

		headersMsg.AddCFHash(filterHash)
	}

	headersMsg.FilterType = msg.FilterType
	headersMsg.StopHash = msg.StopHash
	sp.QueueMessage(headersMsg, nil)
}
