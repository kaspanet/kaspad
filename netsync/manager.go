// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package netsync

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/mempool"
	peerpkg "github.com/kaspanet/kaspad/peer"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

const (
	// maxRejectedTxns is the maximum number of rejected transactions
	// hashes to store in memory.
	maxRejectedTxns = 1000

	// maxRequestedBlocks is the maximum number of requested block
	// hashes to store in memory.
	maxRequestedBlocks = wire.MaxInvPerMsg

	// maxRequestedTxns is the maximum number of requested transactions
	// hashes to store in memory.
	maxRequestedTxns = wire.MaxInvPerMsg

	minGetSelectedTipInterval = time.Minute

	minDAGTimeDelay = time.Minute
)

// newPeerMsg signifies a newly connected peer to the block handler.
type newPeerMsg struct {
	peer *peerpkg.Peer
}

// blockMsg packages a kaspa block message and the peer it came from together
// so the block handler has access to that information.
type blockMsg struct {
	block          *util.Block
	peer           *peerpkg.Peer
	isDelayedBlock bool
	reply          chan struct{}
}

// invMsg packages a kaspa inv message and the peer it came from together
// so the block handler has access to that information.
type invMsg struct {
	inv  *wire.MsgInv
	peer *peerpkg.Peer
}

// donePeerMsg signifies a newly disconnected peer to the block handler.
type donePeerMsg struct {
	peer *peerpkg.Peer
}

// removeFromSyncCandidatesMsg signifies to remove the given peer
// from the sync candidates.
type removeFromSyncCandidatesMsg struct {
	peer *peerpkg.Peer
}

// txMsg packages a kaspa tx message and the peer it came from together
// so the block handler has access to that information.
type txMsg struct {
	tx    *util.Tx
	peer  *peerpkg.Peer
	reply chan struct{}
}

// getSyncPeerMsg is a message type to be sent across the message channel for
// retrieving the current sync peer.
type getSyncPeerMsg struct {
	reply chan int32
}

// processBlockResponse is a response sent to the reply channel of a
// processBlockMsg.
type processBlockResponse struct {
	isOrphan bool
	err      error
}

// processBlockMsg is a message type to be sent across the message channel
// for requested a block is processed. Note this call differs from blockMsg
// above in that blockMsg is intended for blocks that came from peers and have
// extra handling whereas this message essentially is just a concurrent safe
// way to call ProcessBlock on the internal block DAG instance.
type processBlockMsg struct {
	block *util.Block
	flags blockdag.BehaviorFlags
	reply chan processBlockResponse
}

// isCurrentMsg is a message type to be sent across the message channel for
// requesting whether or not the sync manager believes it is synced with the
// currently connected peers.
type isCurrentMsg struct {
	reply chan bool
}

// pauseMsg is a message type to be sent across the message channel for
// pausing the sync manager. This effectively provides the caller with
// exclusive access over the manager until a receive is performed on the
// unpause channel.
type pauseMsg struct {
	unpause <-chan struct{}
}

type selectedTipMsg struct {
	selectedTipHash *daghash.Hash
	peer            *peerpkg.Peer
	reply           chan struct{}
}

type requestQueueAndSet struct {
	queue []*wire.InvVect
	set   map[daghash.Hash]struct{}
}

// peerSyncState stores additional information that the SyncManager tracks
// about a peer.
type peerSyncState struct {
	syncCandidate           bool
	lastSelectedTipRequest  time.Time
	isPendingForSelectedTip bool
	requestQueueMtx         sync.Mutex
	requestQueues           map[wire.InvType]*requestQueueAndSet
	requestedTxns           map[daghash.TxID]struct{}
	requestedBlocks         map[daghash.Hash]struct{}
}

// SyncManager is used to communicate block related messages with peers. The
// SyncManager is started as by executing Start() in a goroutine. Once started,
// it selects peers to sync from and starts the initial block download. Once the
// DAG is in sync, the SyncManager handles incoming block and header
// notifications and relays announcements of new blocks to peers.
type SyncManager struct {
	peerNotifier   PeerNotifier
	started        int32
	shutdown       int32
	dag            *blockdag.BlockDAG
	txMemPool      *mempool.TxPool
	dagParams      *dagconfig.Params
	progressLogger *blockProgressLogger
	msgChan        chan interface{}
	wg             sync.WaitGroup
	quit           chan struct{}

	// These fields should only be accessed from the messageHandler thread
	rejectedTxns    map[daghash.TxID]struct{}
	requestedTxns   map[daghash.TxID]struct{}
	requestedBlocks map[daghash.Hash]struct{}
	syncPeer        *peerpkg.Peer
	peerStates      map[*peerpkg.Peer]*peerSyncState
}

// startSync will choose the sync peer among the available candidate peers to
// download/sync the blockDAG from. When syncing is already running, it
// simply returns. It also examines the candidates for any which are no longer
// candidates and removes them as needed.
func (sm *SyncManager) startSync() {
	// Return now if we're already syncing.
	if sm.syncPeer != nil {
		return
	}

	var syncPeer *peerpkg.Peer
	for peer, state := range sm.peerStates {
		if !state.syncCandidate {
			continue
		}

		if !peer.IsSelectedTipKnown() {
			continue
		}

		// TODO(davec): Use a better algorithm to choose the sync peer.
		// For now, just pick the first available candidate.
		syncPeer = peer
	}

	// Start syncing from the sync peer if one was selected.
	if syncPeer != nil {
		// Clear the requestedBlocks if the sync peer changes, otherwise
		// we may ignore blocks we need that the last sync peer failed
		// to send.
		sm.requestedBlocks = make(map[daghash.Hash]struct{})

		log.Infof("Syncing to block %s from peer %s",
			syncPeer.SelectedTipHash(), syncPeer.Addr())

		syncPeer.PushGetBlockLocatorMsg(syncPeer.SelectedTipHash(), sm.dagParams.GenesisHash)
		sm.syncPeer = syncPeer
		return
	}

	if sm.shouldQueryPeerSelectedTips() {
		hasSyncCandidates := false
		for peer, state := range sm.peerStates {
			if !state.syncCandidate {
				continue
			}
			hasSyncCandidates = true

			if time.Since(state.lastSelectedTipRequest) < minGetSelectedTipInterval {
				continue
			}

			queueMsgGetSelectedTip(peer, state)
		}
		if !hasSyncCandidates {
			log.Warnf("No sync peer candidates available")
		}
	}
}

func (sm *SyncManager) shouldQueryPeerSelectedTips() bool {
	return sm.dag.Now().Sub(sm.dag.CalcPastMedianTime()) > minDAGTimeDelay
}

func queueMsgGetSelectedTip(peer *peerpkg.Peer, state *peerSyncState) {
	state.lastSelectedTipRequest = time.Now()
	state.isPendingForSelectedTip = true
	peer.QueueMessage(wire.NewMsgGetSelectedTip(), nil)
}

// isSyncCandidate returns whether or not the peer is a candidate to consider
// syncing from.
func (sm *SyncManager) isSyncCandidate(peer *peerpkg.Peer) bool {
	// Typically a peer is not a candidate for sync if it's not a full node,
	// however regression test is special in that the regression tool is
	// not a full node and still needs to be considered a sync candidate.
	if sm.dagParams == &dagconfig.RegressionNetParams {
		// The peer is not a candidate if it's not coming from localhost
		// or the hostname can't be determined for some reason.
		host, _, err := net.SplitHostPort(peer.Addr())
		if err != nil {
			return false
		}

		if host != "127.0.0.1" && host != "localhost" {
			return false
		}
	} else {
		// The peer is not a candidate for sync if it's not a full
		// node.
		nodeServices := peer.Services()
		if nodeServices&wire.SFNodeNetwork != wire.SFNodeNetwork {
			return false
		}
	}

	// Candidate if all checks passed.
	return true
}

// handleNewPeerMsg deals with new peers that have signalled they may
// be considered as a sync peer (they have already successfully negotiated). It
// also starts syncing if needed. It is invoked from the syncHandler goroutine.
func (sm *SyncManager) handleNewPeerMsg(peer *peerpkg.Peer) {
	// Ignore if in the process of shutting down.
	if atomic.LoadInt32(&sm.shutdown) != 0 {
		return
	}

	log.Infof("New valid peer %s (%s)", peer, peer.UserAgent())

	// Initialize the peer state
	isSyncCandidate := sm.isSyncCandidate(peer)
	requestQueues := make(map[wire.InvType]*requestQueueAndSet)
	requestQueueInvTypes := []wire.InvType{wire.InvTypeTx, wire.InvTypeBlock, wire.InvTypeSyncBlock}
	for _, invType := range requestQueueInvTypes {
		requestQueues[invType] = &requestQueueAndSet{
			set: make(map[daghash.Hash]struct{}),
		}
	}
	sm.peerStates[peer] = &peerSyncState{
		syncCandidate:   isSyncCandidate,
		requestedTxns:   make(map[daghash.TxID]struct{}),
		requestedBlocks: make(map[daghash.Hash]struct{}),
		requestQueues:   requestQueues,
	}

	// Start syncing by choosing the best candidate if needed.
	if isSyncCandidate && sm.syncPeer == nil {
		sm.startSync()
	}
}

// handleDonePeerMsg deals with peers that have signalled they are done. It
// removes the peer as a candidate for syncing and in the case where it was
// the current sync peer, attempts to select a new best peer to sync from. It
// is invoked from the syncHandler goroutine.
func (sm *SyncManager) handleDonePeerMsg(peer *peerpkg.Peer) {
	state, exists := sm.peerStates[peer]
	if !exists {
		log.Warnf("Received done peer message for unknown peer %s", peer)
		return
	}

	// Remove the peer from the list of candidate peers.
	delete(sm.peerStates, peer)

	log.Infof("Lost peer %s", peer)

	// Remove requested transactions from the global map so that they will
	// be fetched from elsewhere next time we get an inv.
	for txHash := range state.requestedTxns {
		delete(sm.requestedTxns, txHash)
	}

	// Remove requested blocks from the global map so that they will be
	// fetched from elsewhere next time we get an inv.
	// TODO: we could possibly here check which peers have these blocks
	// and request them now to speed things up a little.
	for blockHash := range state.requestedBlocks {
		delete(sm.requestedBlocks, blockHash)
	}

	sm.stopSyncFromPeer(peer)
}

func (sm *SyncManager) stopSyncFromPeer(peer *peerpkg.Peer) {
	// Attempt to find a new peer to sync from if the quitting peer is the
	// sync peer.
	if sm.syncPeer == peer {
		sm.syncPeer = nil
		sm.startSync()
	}
}

func (sm *SyncManager) handleRemoveFromSyncCandidatesMsg(peer *peerpkg.Peer) {
	sm.peerStates[peer].syncCandidate = false
	sm.stopSyncFromPeer(peer)
}

// handleTxMsg handles transaction messages from all peers.
func (sm *SyncManager) handleTxMsg(tmsg *txMsg) {
	peer := tmsg.peer
	state, exists := sm.peerStates[peer]
	if !exists {
		log.Warnf("Received tx message from unknown peer %s", peer)
		return
	}

	// If we didn't ask for this transaction then the peer is misbehaving.
	txID := tmsg.tx.ID()
	if _, exists = state.requestedTxns[*txID]; !exists {
		log.Warnf("Got unrequested transaction %s from %s -- "+
			"disconnecting", txID, peer.Addr())
		peer.Disconnect()
		return
	}

	// Ignore transactions that we have already rejected. Do not
	// send a reject message here because if the transaction was already
	// rejected, the transaction was unsolicited.
	if _, exists = sm.rejectedTxns[*txID]; exists {
		log.Debugf("Ignoring unsolicited previously rejected "+
			"transaction %s from %s", txID, peer)
		return
	}

	// Process the transaction to include validation, insertion in the
	// memory pool, orphan handling, etc.
	acceptedTxs, err := sm.txMemPool.ProcessTransaction(tmsg.tx,
		true, mempool.Tag(peer.ID()))

	// Remove transaction from request maps. Either the mempool/DAG
	// already knows about it and as such we shouldn't have any more
	// instances of trying to fetch it, or we failed to insert and thus
	// we'll retry next time we get an inv.
	delete(state.requestedTxns, *txID)
	delete(sm.requestedTxns, *txID)

	if err != nil {
		// Do not request this transaction again until a new block
		// has been processed.
		sm.rejectedTxns[*txID] = struct{}{}
		sm.limitTxIDMap(sm.rejectedTxns, maxRejectedTxns)

		// When the error is a rule error, it means the transaction was
		// simply rejected as opposed to something actually going wrong,
		// so log it as such. Otherwise, something really did go wrong,
		// so log it as an actual error.
		if errors.As(err, &mempool.RuleError{}) {
			log.Debugf("Rejected transaction %s from %s: %s",
				txID, peer, err)
		} else {
			log.Errorf("Failed to process transaction %s: %s",
				txID, err)
		}

		// Convert the error into an appropriate reject message and
		// send it.
		code, reason := mempool.ErrToRejectErr(err)
		peer.PushRejectMsg(wire.CmdTx, code, reason, (*daghash.Hash)(txID), false)
		return
	}

	sm.peerNotifier.AnnounceNewTransactions(acceptedTxs)
}

// current returns true if we believe we are synced with our peers, false if we
// still have blocks to check
//
// We consider ourselves current iff both of the following are true:
// 1. there's no syncPeer, a.k.a. all connected peers are at the same tip
// 2. the DAG considers itself current - to prevent attacks where a peer sends an
//    unknown tip but never lets us sync to it.
func (sm *SyncManager) current() bool {
	return sm.syncPeer == nil && sm.dag.IsCurrent()
}

// restartSyncIfNeeded finds a new sync candidate if we're not expecting any
// blocks from the current one.
func (sm *SyncManager) restartSyncIfNeeded() {
	if sm.syncPeer != nil {
		syncPeerState, exists := sm.peerStates[sm.syncPeer]
		if exists {
			isWaitingForBlocks := func() bool {
				syncPeerState.requestQueueMtx.Lock()
				defer syncPeerState.requestQueueMtx.Unlock()
				return len(syncPeerState.requestedBlocks) != 0 || len(syncPeerState.requestQueues[wire.InvTypeSyncBlock].queue) != 0
			}()
			if isWaitingForBlocks {
				return
			}
		}
	}

	sm.syncPeer = nil
	sm.startSync()
}

// handleBlockMsg handles block messages from all peers.
func (sm *SyncManager) handleBlockMsg(bmsg *blockMsg) {
	peer := bmsg.peer
	state, exists := sm.peerStates[peer]
	if !exists {
		log.Warnf("Received block message from unknown peer %s", peer)
		return
	}

	// If we didn't ask for this block then the peer is misbehaving.
	blockHash := bmsg.block.Hash()
	if _, exists = state.requestedBlocks[*blockHash]; !exists {
		// The regression test intentionally sends some blocks twice
		// to test duplicate block insertion fails. Don't disconnect
		// the peer or ignore the block when we're in regression test
		// mode in this case so the DAG code is actually fed the
		// duplicate blocks.
		if sm.dagParams != &dagconfig.RegressionNetParams {
			log.Warnf("Got unrequested block %s from %s -- "+
				"disconnecting", blockHash, peer.Addr())
			peer.Disconnect()
			return
		}
	}

	behaviorFlags := blockdag.BFNone
	if bmsg.isDelayedBlock {
		behaviorFlags |= blockdag.BFAfterDelay
	}
	if bmsg.peer == sm.syncPeer {
		behaviorFlags |= blockdag.BFIsSync
	}

	// Process the block to include validation, orphan handling, etc.
	isOrphan, isDelayed, err := sm.dag.ProcessBlock(bmsg.block, behaviorFlags)

	// Remove block from request maps. Either DAG knows about it and
	// so we shouldn't have any more instances of trying to fetch it, or
	// the insertion fails and thus we'll retry next time we get an inv.
	delete(state.requestedBlocks, *blockHash)
	delete(sm.requestedBlocks, *blockHash)

	sm.restartSyncIfNeeded()

	if err != nil {
		// When the error is a rule error, it means the block was simply
		// rejected as opposed to something actually going wrong, so log
		// it as such. Otherwise, something really did go wrong, so log
		// it as an actual error.
		if !errors.As(err, &blockdag.RuleError{}) {
			panic(errors.Wrapf(err, "Failed to process block %s",
				blockHash))
		}
		log.Infof("Rejected block %s from %s: %s", blockHash,
			peer, err)

		// Convert the error into an appropriate reject message and
		// send it.
		code, reason := mempool.ErrToRejectErr(err)
		peer.PushRejectMsg(wire.CmdBlock, code, reason, blockHash, false)

		// Disconnect from the misbehaving peer.
		peer.Disconnect()
		return
	}

	if isDelayed {
		return
	}

	// Request the parents for the orphan block from the peer that sent it.
	if isOrphan {
		missingAncestors, err := sm.dag.GetOrphanMissingAncestorHashes(blockHash)
		if err != nil {
			log.Errorf("Failed to find missing ancestors for block %s: %s",
				blockHash, err)
			return
		}
		sm.addBlocksToRequestQueue(state, missingAncestors, false)
	} else {
		// When the block is not an orphan, log information about it and
		// update the DAG state.
		blockBlueScore, err := sm.dag.BlueScoreByBlockHash(blockHash)
		if err != nil {
			log.Errorf("Failed to get blue score for block %s: %s", blockHash, err)
		}
		sm.progressLogger.LogBlockBlueScore(bmsg.block, blockBlueScore)

		// Clear the rejected transactions.
		sm.rejectedTxns = make(map[daghash.TxID]struct{})
	}

	// We don't want to flood our sync peer with getdata messages, so
	// instead of asking it immediately about missing ancestors, we first
	// wait until it finishes to send us all of the requested blocks.
	if (isOrphan && peer != sm.syncPeer) || (peer == sm.syncPeer && len(state.requestedBlocks) == 0) {
		err := sm.sendInvsFromRequestQueue(peer, state)
		if err != nil {
			log.Errorf("Failed to send invs from queue: %s", err)
			return
		}
	}
}

func (sm *SyncManager) addBlocksToRequestQueue(state *peerSyncState, hashes []*daghash.Hash, isRelayedInv bool) {
	state.requestQueueMtx.Lock()
	defer state.requestQueueMtx.Unlock()
	for _, hash := range hashes {
		if _, exists := sm.requestedBlocks[*hash]; !exists {
			invType := wire.InvTypeSyncBlock
			if isRelayedInv {
				invType = wire.InvTypeBlock
			}
			iv := wire.NewInvVect(invType, hash)
			state.addInvToRequestQueueNoLock(iv)
		}
	}
}

func (state *peerSyncState) addInvToRequestQueueNoLock(iv *wire.InvVect) {
	requestQueue, ok := state.requestQueues[iv.Type]
	if !ok {
		panic(errors.Errorf("got unsupported inventory type %s", iv.Type))
	}
	if _, exists := requestQueue.set[*iv.Hash]; !exists {
		requestQueue.set[*iv.Hash] = struct{}{}
		requestQueue.queue = append(requestQueue.queue, iv)
	}
}

func (state *peerSyncState) addInvToRequestQueue(iv *wire.InvVect) {
	state.requestQueueMtx.Lock()
	defer state.requestQueueMtx.Unlock()
	state.addInvToRequestQueueNoLock(iv)
}

// haveInventory returns whether or not the inventory represented by the passed
// inventory vector is known. This includes checking all of the various places
// inventory can be when it is in different states such as blocks that are part
// of the DAG, in the orphan pool, and transactions that are in the memory pool
// (either the main pool or orphan pool).
func (sm *SyncManager) haveInventory(invVect *wire.InvVect) (bool, error) {
	switch invVect.Type {
	case wire.InvTypeSyncBlock:
		fallthrough
	case wire.InvTypeBlock:
		// Ask DAG if the block is known to it in any form (in DAG or as an orphan).
		return sm.dag.IsKnownBlock(invVect.Hash), nil

	case wire.InvTypeTx:
		// Ask the transaction memory pool if the transaction is known
		// to it in any form (main pool or orphan).
		if sm.txMemPool.HaveTransaction((*daghash.TxID)(invVect.Hash)) {
			return true, nil
		}

		// Check if the transaction exists from the point of view of the
		// DAG's virtual block. Note that this is only a best effort
		// since it is expensive to check existence of every output and
		// the only purpose of this check is to avoid downloading
		// already known transactions. Only the first two outputs are
		// checked because the vast majority of transactions consist of
		// two outputs where one is some form of "pay-to-somebody-else"
		// and the other is a change output.
		prevOut := wire.Outpoint{TxID: daghash.TxID(*invVect.Hash)}
		for i := uint32(0); i < 2; i++ {
			prevOut.Index = i
			entry, ok := sm.dag.GetUTXOEntry(prevOut)
			if !ok {
				return false, nil
			}
			if entry != nil {
				return true, nil
			}
		}

		return false, nil
	}

	// The requested inventory is is an unsupported type, so just claim
	// it is known to avoid requesting it.
	return true, nil
}

// handleInvMsg handles inv messages from all peers.
// We examine the inventory advertised by the remote peer and act accordingly.
func (sm *SyncManager) handleInvMsg(imsg *invMsg) {
	peer := imsg.peer
	state, exists := sm.peerStates[peer]
	if !exists {
		log.Warnf("Received inv message from unknown peer %s", peer)
		return
	}

	// Attempt to find the final block in the inventory list. There may
	// not be one.
	lastBlock := -1
	invVects := imsg.inv.InvList
	for i := len(invVects) - 1; i >= 0; i-- {
		if invVects[i].IsBlockOrSyncBlock() {
			lastBlock = i
			break
		}
	}

	haveUnknownInvBlock := false

	// Request the advertised inventory if we don't already have it. Also,
	// request parent blocks of orphans if we receive one we already have.
	// Finally, attempt to detect potential stalls due to big orphan DAGs
	// we already have and request more blocks to prevent them.
	for i, iv := range invVects {
		// Ignore unsupported inventory types.
		switch iv.Type {
		case wire.InvTypeBlock:
		case wire.InvTypeSyncBlock:
		case wire.InvTypeTx:
		default:
			continue
		}

		// Add the inventory to the cache of known inventory
		// for the peer.
		peer.AddKnownInventory(iv)

		// Request the inventory if we don't already have it.
		haveInv, err := sm.haveInventory(iv)
		if err != nil {
			log.Warnf("Unexpected failure when checking for "+
				"existing inventory during inv message "+
				"processing: %s", err)
			continue
		}
		if !haveInv {
			if iv.Type == wire.InvTypeTx {
				// Skip the transaction if it has already been rejected.
				if _, exists := sm.rejectedTxns[daghash.TxID(*iv.Hash)]; exists {
					continue
				}

				// Skip the transaction if it had previously been requested.
				if _, exists := state.requestedTxns[daghash.TxID(*iv.Hash)]; exists {
					continue
				}
			}

			if iv.Type == wire.InvTypeBlock {
				haveUnknownInvBlock = true
			}

			// Add it to the request queue.
			state.addInvToRequestQueue(iv)
			continue
		}

		if iv.IsBlockOrSyncBlock() {
			// The block is an orphan block that we already have.
			// When the existing orphan was processed, it requested
			// the missing parent blocks. When this scenario
			// happens, it means there were more blocks missing
			// than are allowed into a single inventory message. As
			// a result, once this peer requested the final
			// advertised block, the remote peer noticed and is now
			// resending the orphan block as an available block
			// to signal there are more missing blocks that need to
			// be requested.
			if sm.dag.IsKnownOrphan(iv.Hash) {
				missingAncestors, err := sm.dag.GetOrphanMissingAncestorHashes(iv.Hash)
				if err != nil {
					log.Errorf("Failed to find missing ancestors for block %s: %s",
						iv.Hash, err)
					return
				}
				sm.addBlocksToRequestQueue(state, missingAncestors, iv.Type != wire.InvTypeSyncBlock)
				continue
			}

			// We already have the final block advertised by this
			// inventory message, so force a request for more. This
			// should only happen if our DAG and the peer's DAG have
			// diverged long time ago.
			if i == lastBlock && peer == sm.syncPeer {
				// Request blocks after the first block's ancestor that exists
				// in the selected path chain, one up to the
				// final one the remote peer knows about.
				peer.PushGetBlockLocatorMsg(iv.Hash, sm.dagParams.GenesisHash)
			}
		}
	}

	err := sm.sendInvsFromRequestQueue(peer, state)
	if err != nil {
		log.Errorf("Failed to send invs from queue: %s", err)
	}

	if haveUnknownInvBlock && !sm.current() {
		// If one of the inv messages is an unknown block
		// it is an indication that one of our peers has more
		// up-to-date data than us.
		sm.restartSyncIfNeeded()
	}
}

func (sm *SyncManager) addInvsToGetDataMessageFromQueue(gdmsg *wire.MsgGetData, state *peerSyncState, invType wire.InvType, maxInvsToAdd int) error {
	requestQueue, ok := state.requestQueues[invType]
	if !ok {
		panic(errors.Errorf("got unsupported inventory type %s", invType))
	}
	queue := requestQueue.queue
	var invsNum int
	leftSpaceInGdmsg := wire.MaxInvPerGetDataMsg - len(gdmsg.InvList)
	if len(queue) > leftSpaceInGdmsg {
		invsNum = leftSpaceInGdmsg
	} else {
		invsNum = len(queue)
	}
	if invsNum > maxInvsToAdd {
		invsNum = maxInvsToAdd
	}
	invsToAdd := make([]*wire.InvVect, 0, invsNum)
	for len(queue) != 0 && len(invsToAdd) < invsNum {
		var iv *wire.InvVect
		iv, queue = queue[0], queue[1:]

		exists, err := sm.haveInventory(iv)
		if err != nil {
			return err
		}
		if !exists {
			invsToAdd = append(invsToAdd, iv)
		}
	}

	addBlockInv := func(iv *wire.InvVect) {
		// Request the block if there is not already a pending
		// request.
		if _, exists := sm.requestedBlocks[*iv.Hash]; !exists {
			sm.requestedBlocks[*iv.Hash] = struct{}{}
			sm.limitHashMap(sm.requestedBlocks, maxRequestedBlocks)
			state.requestedBlocks[*iv.Hash] = struct{}{}

			gdmsg.AddInvVect(iv)
		}
	}
	for _, iv := range invsToAdd {
		delete(requestQueue.set, *iv.Hash)
		switch invType {
		case wire.InvTypeSyncBlock:
			addBlockInv(iv)
		case wire.InvTypeBlock:
			addBlockInv(iv)

		case wire.InvTypeTx:
			// Request the transaction if there is not already a
			// pending request.
			if _, exists := sm.requestedTxns[daghash.TxID(*iv.Hash)]; !exists {
				sm.requestedTxns[daghash.TxID(*iv.Hash)] = struct{}{}
				sm.limitTxIDMap(sm.requestedTxns, maxRequestedTxns)
				state.requestedTxns[daghash.TxID(*iv.Hash)] = struct{}{}

				gdmsg.AddInvVect(iv)
			}
		}

		if len(queue) >= wire.MaxInvPerGetDataMsg {
			break
		}
	}
	requestQueue.queue = queue
	return nil
}

func (sm *SyncManager) sendInvsFromRequestQueue(peer *peerpkg.Peer, state *peerSyncState) error {
	state.requestQueueMtx.Lock()
	defer state.requestQueueMtx.Unlock()
	gdmsg := wire.NewMsgGetData()
	err := sm.addInvsToGetDataMessageFromQueue(gdmsg, state, wire.InvTypeSyncBlock, wire.MaxSyncBlockInvPerGetDataMsg)
	if err != nil {
		return err
	}
	if sm.current() {
		err := sm.addInvsToGetDataMessageFromQueue(gdmsg, state, wire.InvTypeBlock, wire.MaxInvPerGetDataMsg)
		if err != nil {
			return err
		}

		err = sm.addInvsToGetDataMessageFromQueue(gdmsg, state, wire.InvTypeTx, wire.MaxInvPerGetDataMsg)
		if err != nil {
			return err
		}
	}
	if len(gdmsg.InvList) > 0 {
		peer.QueueMessage(gdmsg, nil)
	}
	return nil
}

// limitTxIDMap is a helper function for maps that require a maximum limit by
// evicting a random transaction if adding a new value would cause it to
// overflow the maximum allowed.
func (sm *SyncManager) limitTxIDMap(m map[daghash.TxID]struct{}, limit int) {
	if len(m)+1 > limit {
		// Remove a random entry from the map. For most compilers, Go's
		// range statement iterates starting at a random item although
		// that is not 100% guaranteed by the spec. The iteration order
		// is not important here because an adversary would have to be
		// able to pull off preimage attacks on the hashing function in
		// order to target eviction of specific entries anyways.
		for txID := range m {
			delete(m, txID)
			return
		}
	}
}

// limitHashMap is a helper function for maps that require a maximum limit by
// evicting a random item if adding a new value would cause it to
// overflow the maximum allowed.
func (sm *SyncManager) limitHashMap(m map[daghash.Hash]struct{}, limit int) {
	if len(m)+1 > limit {
		// Remove a random entry from the map. For most compilers, Go's
		// range statement iterates starting at a random item although
		// that is not 100% guaranteed by the spec. The iteration order
		// is not important here because an adversary would have to be
		// able to pull off preimage attacks on the hashing function in
		// order to target eviction of specific entries anyways.
		for hash := range m {
			delete(m, hash)
			return
		}
	}
}

func (sm *SyncManager) handleProcessBlockMsg(msg processBlockMsg) (isOrphan bool, err error) {
	isOrphan, isDelayed, err := sm.dag.ProcessBlock(
		msg.block, msg.flags|blockdag.BFDisallowDelay)
	if err != nil {
		return false, err
	}
	if isDelayed {
		return false, errors.New("Cannot process blocks from RPC beyond the allowed time offset")
	}

	return isOrphan, nil
}

func (sm *SyncManager) handleSelectedTipMsg(msg *selectedTipMsg) {
	peer := msg.peer
	selectedTipHash := msg.selectedTipHash
	state := sm.peerStates[peer]
	if !state.isPendingForSelectedTip {
		log.Warnf("Got unrequested selected tip message from %s -- "+
			"disconnecting", peer.Addr())
		peer.Disconnect()
	}
	state.isPendingForSelectedTip = false
	if selectedTipHash.IsEqual(peer.SelectedTipHash()) {
		return
	}
	peer.SetSelectedTipHash(selectedTipHash)
	sm.startSync()
}

// messageHandler is the main handler for the sync manager. It must be run as a
// goroutine. It processes block and inv messages in a separate goroutine
// from the peer handlers so the block (MsgBlock) messages are handled by a
// single thread without needing to lock memory data structures. This is
// important because the sync manager controls which blocks are needed and how
// the fetching should proceed.
func (sm *SyncManager) messageHandler() {
out:
	for {
		select {
		case m := <-sm.msgChan:
			switch msg := m.(type) {
			case *newPeerMsg:
				sm.handleNewPeerMsg(msg.peer)

			case *txMsg:
				sm.handleTxMsg(msg)
				msg.reply <- struct{}{}

			case *blockMsg:
				sm.handleBlockMsg(msg)
				msg.reply <- struct{}{}

			case *invMsg:
				sm.handleInvMsg(msg)

			case *donePeerMsg:
				sm.handleDonePeerMsg(msg.peer)

			case *removeFromSyncCandidatesMsg:
				sm.handleRemoveFromSyncCandidatesMsg(msg.peer)

			case getSyncPeerMsg:
				var peerID int32
				if sm.syncPeer != nil {
					peerID = sm.syncPeer.ID()
				}
				msg.reply <- peerID

			case processBlockMsg:
				isOrphan, err := sm.handleProcessBlockMsg(msg)
				msg.reply <- processBlockResponse{
					isOrphan: isOrphan,
					err:      err,
				}

			case isCurrentMsg:
				msg.reply <- sm.current()

			case pauseMsg:
				// Wait until the sender unpauses the manager.
				<-msg.unpause

			case *selectedTipMsg:
				sm.handleSelectedTipMsg(msg)
				msg.reply <- struct{}{}

			default:
				log.Warnf("Invalid message type in block "+
					"handler: %T", msg)
			}

		case <-sm.quit:
			break out
		}
	}

	sm.wg.Done()
	log.Trace("Block handler done")
}

// handleBlockDAGNotification handles notifications from blockDAG. It does
// things such as request orphan block parents and relay accepted blocks to
// connected peers.
func (sm *SyncManager) handleBlockDAGNotification(notification *blockdag.Notification) {
	switch notification.Type {
	// A block has been accepted into the blockDAG. Relay it to other peers.
	case blockdag.NTBlockAdded:
		data, ok := notification.Data.(*blockdag.BlockAddedNotificationData)
		if !ok {
			log.Warnf("Block Added notification data is of wrong type.")
			break
		}
		block := data.Block

		// Update mempool
		ch := make(chan mempool.NewBlockMsg)
		spawn(func() {
			err := sm.txMemPool.HandleNewBlock(block, ch)
			close(ch)
			if err != nil {
				panic(fmt.Sprintf("HandleNewBlock failed to handle block %s", block.Hash()))
			}
		})

		// Relay if we are current and the block was not just now unorphaned.
		// Otherwise peers that are current should already know about it
		if sm.current() && !data.WasUnorphaned {
			iv := wire.NewInvVect(wire.InvTypeBlock, block.Hash())
			sm.peerNotifier.RelayInventory(iv, block.MsgBlock().Header)
		}

		for msg := range ch {
			sm.peerNotifier.TransactionConfirmed(msg.Tx)
			sm.peerNotifier.AnnounceNewTransactions(msg.AcceptedTxs)
		}
	}
}

// NewPeer informs the sync manager of a newly active peer.
func (sm *SyncManager) NewPeer(peer *peerpkg.Peer) {
	// Ignore if we are shutting down.
	if atomic.LoadInt32(&sm.shutdown) != 0 {
		return
	}
	sm.msgChan <- &newPeerMsg{peer: peer}
}

// QueueTx adds the passed transaction message and peer to the block handling
// queue. Responds to the done channel argument after the tx message is
// processed.
func (sm *SyncManager) QueueTx(tx *util.Tx, peer *peerpkg.Peer, done chan struct{}) {
	// Don't accept more transactions if we're shutting down.
	if atomic.LoadInt32(&sm.shutdown) != 0 {
		done <- struct{}{}
		return
	}

	sm.msgChan <- &txMsg{tx: tx, peer: peer, reply: done}
}

// QueueBlock adds the passed block message and peer to the block handling
// queue. Responds to the done channel argument after the block message is
// processed.
func (sm *SyncManager) QueueBlock(block *util.Block, peer *peerpkg.Peer, isDelayedBlock bool, done chan struct{}) {
	// Don't accept more blocks if we're shutting down.
	if atomic.LoadInt32(&sm.shutdown) != 0 {
		done <- struct{}{}
		return
	}

	sm.msgChan <- &blockMsg{block: block, peer: peer, isDelayedBlock: isDelayedBlock, reply: done}
}

// QueueInv adds the passed inv message and peer to the block handling queue.
func (sm *SyncManager) QueueInv(inv *wire.MsgInv, peer *peerpkg.Peer) {
	// No channel handling here because peers do not need to block on inv
	// messages.
	if atomic.LoadInt32(&sm.shutdown) != 0 {
		return
	}

	sm.msgChan <- &invMsg{inv: inv, peer: peer}
}

// QueueSelectedTipMsg adds the passed selected tip message and peer to the
// block handling queue. Responds to the done channel argument after it finished
// handling the message.
func (sm *SyncManager) QueueSelectedTipMsg(msg *wire.MsgSelectedTip, peer *peerpkg.Peer, done chan struct{}) {
	sm.msgChan <- &selectedTipMsg{
		selectedTipHash: msg.SelectedTipHash,
		peer:            peer,
		reply:           done,
	}
}

// DonePeer informs the blockmanager that a peer has disconnected.
func (sm *SyncManager) DonePeer(peer *peerpkg.Peer) {
	// Ignore if we are shutting down.
	if atomic.LoadInt32(&sm.shutdown) != 0 {
		return
	}

	sm.msgChan <- &donePeerMsg{peer: peer}
}

// RemoveFromSyncCandidates tells the blockmanager to remove a peer as
// a sync candidate.
func (sm *SyncManager) RemoveFromSyncCandidates(peer *peerpkg.Peer) {
	// Ignore if we are shutting down.
	if atomic.LoadInt32(&sm.shutdown) != 0 {
		return
	}

	sm.msgChan <- &removeFromSyncCandidatesMsg{peer: peer}
}

// Start begins the core block handler which processes block and inv messages.
func (sm *SyncManager) Start() {
	// Already started?
	if atomic.AddInt32(&sm.started, 1) != 1 {
		return
	}

	log.Trace("Starting sync manager")
	sm.wg.Add(1)
	spawn(sm.messageHandler)
}

// Stop gracefully shuts down the sync manager by stopping all asynchronous
// handlers and waiting for them to finish.
func (sm *SyncManager) Stop() error {
	if atomic.AddInt32(&sm.shutdown, 1) != 1 {
		log.Warnf("Sync manager is already in the process of " +
			"shutting down")
		return nil
	}

	log.Infof("Sync manager shutting down")
	close(sm.quit)
	sm.wg.Wait()
	return nil
}

// SyncPeerID returns the ID of the current sync peer, or 0 if there is none.
func (sm *SyncManager) SyncPeerID() int32 {
	reply := make(chan int32)
	sm.msgChan <- getSyncPeerMsg{reply: reply}
	return <-reply
}

// ProcessBlock makes use of ProcessBlock on an internal instance of a blockDAG.
func (sm *SyncManager) ProcessBlock(block *util.Block, flags blockdag.BehaviorFlags) (bool, error) {
	reply := make(chan processBlockResponse)
	sm.msgChan <- processBlockMsg{block: block, flags: flags, reply: reply}
	response := <-reply
	return response.isOrphan, response.err
}

// IsCurrent returns whether or not the sync manager believes it is synced with
// the connected peers.
func (sm *SyncManager) IsCurrent() bool {
	reply := make(chan bool)
	sm.msgChan <- isCurrentMsg{reply: reply}
	return <-reply
}

// Pause pauses the sync manager until the returned channel is closed.
//
// Note that while paused, all peer and block processing is halted. The
// message sender should avoid pausing the sync manager for long durations.
func (sm *SyncManager) Pause() chan<- struct{} {
	c := make(chan struct{})
	sm.msgChan <- pauseMsg{c}
	return c
}

// New constructs a new SyncManager. Use Start to begin processing asynchronous
// block, tx, and inv updates.
func New(config *Config) (*SyncManager, error) {
	sm := SyncManager{
		peerNotifier:    config.PeerNotifier,
		dag:             config.DAG,
		txMemPool:       config.TxMemPool,
		dagParams:       config.DAGParams,
		rejectedTxns:    make(map[daghash.TxID]struct{}),
		requestedTxns:   make(map[daghash.TxID]struct{}),
		requestedBlocks: make(map[daghash.Hash]struct{}),
		peerStates:      make(map[*peerpkg.Peer]*peerSyncState),
		progressLogger:  newBlockProgressLogger("Processed", log),
		msgChan:         make(chan interface{}, config.MaxPeers*3),
		quit:            make(chan struct{}),
	}

	sm.dag.Subscribe(sm.handleBlockDAGNotification)

	return &sm, nil
}
