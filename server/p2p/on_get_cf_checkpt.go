package p2p

import (
	"github.com/kaspanet/kaspad/peer"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

// OnGetCFCheckpt is invoked when a peer receives a getcfcheckpt bitcoin message.
func (sp *Peer) OnGetCFCheckpt(_ *peer.Peer, msg *wire.MsgGetCFCheckpt) {
	// Ignore getcfcheckpt requests if not in sync.
	if !sp.server.SyncManager.IsCurrent() {
		return
	}

	blockHashes, err := sp.server.DAG.IntervalBlockHashes(msg.StopHash,
		wire.CFCheckptInterval)
	if err != nil {
		peerLog.Debugf("Invalid getcfilters request: %s", err)
		return
	}

	var updateCache bool
	var checkptCache []cfHeaderKV

	if len(blockHashes) > len(checkptCache) {
		// Update the cache if the checkpoint chain is longer than the cached
		// one. This ensures that the cache is relatively stable and mostly
		// overlaps with the best chain, since it follows the longest chain
		// heuristic.
		updateCache = true

		// Take write lock because we are going to update cache.
		sp.server.cfCheckptCachesMtx.Lock()
		defer sp.server.cfCheckptCachesMtx.Unlock()

		// Grow the checkptCache to be the length of blockHashes.
		additionalLength := len(blockHashes) - len(checkptCache)
		checkptCache = append(sp.server.cfCheckptCaches[msg.FilterType],
			make([]cfHeaderKV, additionalLength)...)
	} else {
		updateCache = false

		// Take reader lock because we are not going to update cache.
		sp.server.cfCheckptCachesMtx.RLock()
		defer sp.server.cfCheckptCachesMtx.RUnlock()

		checkptCache = sp.server.cfCheckptCaches[msg.FilterType]
	}

	// Iterate backwards until the block hash is found in the cache.
	var forkIdx int
	for forkIdx = len(checkptCache); forkIdx > 0; forkIdx-- {
		if checkptCache[forkIdx-1].blockHash.IsEqual(blockHashes[forkIdx-1]) {
			break
		}
	}

	// Populate results with cached checkpoints.
	checkptMsg := wire.NewMsgCFCheckpt(msg.FilterType, msg.StopHash,
		len(blockHashes))
	for i := 0; i < forkIdx; i++ {
		checkptMsg.AddCFHeader(checkptCache[i].filterHeader)
	}

	// Look up any filter headers that aren't cached.
	blockHashPtrs := make([]*daghash.Hash, 0, len(blockHashes)-forkIdx)
	for i := forkIdx; i < len(blockHashes); i++ {
		blockHashPtrs = append(blockHashPtrs, blockHashes[i])
	}

	filterHeaders, err := sp.server.CfIndex.FilterHeadersByBlockHashes(blockHashPtrs,
		msg.FilterType)
	if err != nil {
		peerLog.Errorf("Error retrieving cfilter headers: %s", err)
		return
	}

	for i, filterHeaderBytes := range filterHeaders {
		if len(filterHeaderBytes) == 0 {
			peerLog.Warnf("Could not obtain CF header for %s", blockHashPtrs[i])
			return
		}

		filterHeader, err := daghash.NewHash(filterHeaderBytes)
		if err != nil {
			peerLog.Warnf("Committed filter header deserialize "+
				"failed: %s", err)
			return
		}

		checkptMsg.AddCFHeader(filterHeader)
		if updateCache {
			checkptCache[forkIdx+i] = cfHeaderKV{
				blockHash:    blockHashes[forkIdx+i],
				filterHeader: filterHeader,
			}
		}
	}

	if updateCache {
		sp.server.cfCheckptCaches[msg.FilterType] = checkptCache
	}

	sp.QueueMessage(checkptMsg, nil)
}
