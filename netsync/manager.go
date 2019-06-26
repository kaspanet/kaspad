// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package netsync

import (
	"container/list"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/mempool"
	peerpkg "github.com/daglabs/btcd/peer"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/wire"
)

const (
	// minInFlightBlocks is the minimum number of blocks that should be
	// in the request queue for headers-first mode before requesting
	// more.
	minInFlightBlocks = 10

	// maxRejectedTxns is the maximum number of rejected transactions
	// hashes to store in memory.
	maxRejectedTxns = 1000

	// maxRequestedBlocks is the maximum number of requested block
	// hashes to store in memory.
	maxRequestedBlocks = wire.MaxInvPerMsg

	// maxRequestedTxns is the maximum number of requested transactions
	// hashes to store in memory.
	maxRequestedTxns = wire.MaxInvPerMsg
)

// newPeerMsg signifies a newly connected peer to the block handler.
type newPeerMsg struct {
	peer *peerpkg.Peer
}

// blockMsg packages a bitcoin block message and the peer it came from together
// so the block handler has access to that information.
type blockMsg struct {
	block          *util.Block
	peer           *peerpkg.Peer
	isDelayedBlock bool
	reply          chan struct{}
}

// invMsg packages a bitcoin inv message and the peer it came from together
// so the block handler has access to that information.
type invMsg struct {
	inv  *wire.MsgInv
	peer *peerpkg.Peer
}

// headersMsg packages a bitcoin headers message and the peer it came from
// together so the block handler has access to that information.
type headersMsg struct {
	headers *wire.MsgHeaders
	peer    *peerpkg.Peer
}

// donePeerMsg signifies a newly disconnected peer to the block handler.
type donePeerMsg struct {
	peer *peerpkg.Peer
}

// txMsg packages a bitcoin tx message and the peer it came from together
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
// for requested a block is processed.  Note this call differs from blockMsg
// above in that blockMsg is intended for blocks that came from peers and have
// extra handling whereas this message essentially is just a concurrent safe
// way to call ProcessBlock on the internal block chain instance.
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
// pausing the sync manager.  This effectively provides the caller with
// exclusive access over the manager until a receive is performed on the
// unpause channel.
type pauseMsg struct {
	unpause <-chan struct{}
}

// headerNode is used as a node in a list of headers that are linked together
// between checkpoints.
type headerNode struct {
	height uint64
	hash   *daghash.Hash
}

// peerSyncState stores additional information that the SyncManager tracks
// about a peer.
type peerSyncState struct {
	syncCandidate   bool
	requestQueueMtx sync.Mutex
	// relayedInvsRequestQueue contains invs of blocks and transactions
	// which are relayed to our node.
	relayedInvsRequestQueue []*wire.InvVect
	// requestQueue contains all of the invs that are not relayed to
	// us; we get them by requesting them or by manually creating them.
	requestQueue               []*wire.InvVect
	relayedInvsRequestQueueSet map[daghash.Hash]struct{}
	requestQueueSet            map[daghash.Hash]struct{}
	requestedTxns              map[daghash.TxID]struct{}
	requestedBlocks            map[daghash.Hash]struct{}
}

// SyncManager is used to communicate block related messages with peers. The
// SyncManager is started as by executing Start() in a goroutine. Once started,
// it selects peers to sync from and starts the initial block download. Once the
// chain is in sync, the SyncManager handles incoming block and header
// notifications and relays announcements of new blocks to peers.
type SyncManager struct {
	peerNotifier   PeerNotifier
	started        int32
	shutdown       int32
	dag            *blockdag.BlockDAG
	txMemPool      *mempool.TxPool
	chainParams    *dagconfig.Params
	progressLogger *blockProgressLogger
	msgChan        chan interface{}
	wg             sync.WaitGroup
	quit           chan struct{}

	// These fields should only be accessed from the blockHandler thread
	rejectedTxns    map[daghash.TxID]struct{}
	requestedTxns   map[daghash.TxID]struct{}
	requestedBlocks map[daghash.Hash]struct{}
	syncPeer        *peerpkg.Peer
	peerStates      map[*peerpkg.Peer]*peerSyncState

	// The following fields are used for headers-first mode.
	headersFirstMode bool
	headerList       *list.List
	startHeader      *list.Element
	nextCheckpoint   *dagconfig.Checkpoint
}

// resetHeaderState sets the headers-first mode state to values appropriate for
// syncing from a new peer.
func (sm *SyncManager) resetHeaderState(newestHash *daghash.Hash, newestHeight uint64) {
	sm.headersFirstMode = false
	sm.headerList.Init()
	sm.startHeader = nil

	// When there is a next checkpoint, add an entry for the latest known
	// block into the header pool.  This allows the next downloaded header
	// to prove it links to the chain properly.
	if sm.nextCheckpoint != nil {
		node := headerNode{height: newestHeight, hash: newestHash}
		sm.headerList.PushBack(&node)
	}
}

// findNextHeaderCheckpoint returns the next checkpoint after the passed height.
// It returns nil when there is not one either because the height is already
// later than the final checkpoint or some other reason such as disabled
// checkpoints.
func (sm *SyncManager) findNextHeaderCheckpoint(height uint64) *dagconfig.Checkpoint {
	checkpoints := sm.dag.Checkpoints()
	if len(checkpoints) == 0 {
		return nil
	}

	// There is no next checkpoint if the height is already after the final
	// checkpoint.
	finalCheckpoint := &checkpoints[len(checkpoints)-1]
	if height >= finalCheckpoint.ChainHeight {
		return nil
	}

	// Find the next checkpoint.
	nextCheckpoint := finalCheckpoint
	for i := len(checkpoints) - 2; i >= 0; i-- {
		if height >= checkpoints[i].ChainHeight {
			break
		}
		nextCheckpoint = &checkpoints[i]
	}
	return nextCheckpoint
}

// startSync will choose the best peer among the available candidate peers to
// download/sync the blockchain from.  When syncing is already running, it
// simply returns.  It also examines the candidates for any which are no longer
// candidates and removes them as needed.
func (sm *SyncManager) startSync() {
	// Return now if we're already syncing.
	if sm.syncPeer != nil {
		return
	}

	var bestPeer *peerpkg.Peer
	for peer, state := range sm.peerStates {
		if !state.syncCandidate {
			continue
		}

		isCandidate, err := peer.IsSyncCandidate()
		if err != nil {
			log.Errorf("Failed to check if peer %s is"+
				"a sync candidate: %s", peer, err)
			return
		}

		if !isCandidate {
			state.syncCandidate = false
			continue
		}

		// TODO(davec): Use a better algorithm to choose the best peer.
		// For now, just pick the first available candidate.
		bestPeer = peer
	}

	// Start syncing from the best peer if one was selected.
	if bestPeer != nil {
		// Clear the requestedBlocks if the sync peer changes, otherwise
		// we may ignore blocks we need that the last sync peer failed
		// to send.
		sm.requestedBlocks = make(map[daghash.Hash]struct{})

		locator := sm.dag.LatestBlockLocator()

		log.Infof("Syncing to block %s from peer %s",
			bestPeer.SelectedTip(), bestPeer.Addr())

		// When the current height is less than a known checkpoint we
		// can use block headers to learn about which blocks comprise
		// the chain up to the checkpoint and perform less validation
		// for them.  This is possible since each header contains the
		// hash of the previous header and a merkle root.  Therefore if
		// we validate all of the received headers link together
		// properly and the checkpoint hashes match, we can be sure the
		// hashes for the blocks in between are accurate.  Further, once
		// the full blocks are downloaded, the merkle root is computed
		// and compared against the value in the header which proves the
		// full block hasn't been tampered with.
		//
		// Once we have passed the final checkpoint, or checkpoints are
		// disabled, use standard inv messages learn about the blocks
		// and fully validate them.  Finally, regression test mode does
		// not support the headers-first approach so do normal block
		// downloads when in regression test mode.
		if sm.nextCheckpoint != nil &&
			sm.dag.ChainHeight() < sm.nextCheckpoint.ChainHeight &&
			sm.chainParams != &dagconfig.RegressionNetParams { //TODO: (Ori) This is probably wrong. Done only for compilation

			bestPeer.PushGetHeadersMsg(locator, sm.nextCheckpoint.Hash)
			sm.headersFirstMode = true
			log.Infof("Downloading headers for blocks %d to "+
				"%d from peer %s", sm.dag.ChainHeight()+1,
				sm.nextCheckpoint.ChainHeight, bestPeer.Addr()) //TODO: (Ori) This is probably wrong. Done only for compilation
		} else {
			bestPeer.PushGetBlocksMsg(locator, &daghash.ZeroHash)
		}
		sm.syncPeer = bestPeer
	} else {
		log.Warnf("No sync peer candidates available")
	}
}

// isSyncCandidate returns whether or not the peer is a candidate to consider
// syncing from.
func (sm *SyncManager) isSyncCandidate(peer *peerpkg.Peer) bool {
	// Typically a peer is not a candidate for sync if it's not a full node,
	// however regression test is special in that the regression tool is
	// not a full node and still needs to be considered a sync candidate.
	if sm.chainParams == &dagconfig.RegressionNetParams {
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
// be considered as a sync peer (they have already successfully negotiated).  It
// also starts syncing if needed.  It is invoked from the syncHandler goroutine.
func (sm *SyncManager) handleNewPeerMsg(peer *peerpkg.Peer) {
	// Ignore if in the process of shutting down.
	if atomic.LoadInt32(&sm.shutdown) != 0 {
		return
	}

	log.Infof("New valid peer %s (%s)", peer, peer.UserAgent())

	// Initialize the peer state
	isSyncCandidate := sm.isSyncCandidate(peer)
	sm.peerStates[peer] = &peerSyncState{
		syncCandidate:              isSyncCandidate,
		requestedTxns:              make(map[daghash.TxID]struct{}),
		requestedBlocks:            make(map[daghash.Hash]struct{}),
		requestQueueSet:            make(map[daghash.Hash]struct{}),
		relayedInvsRequestQueueSet: make(map[daghash.Hash]struct{}),
	}

	// Start syncing by choosing the best candidate if needed.
	if isSyncCandidate && sm.syncPeer == nil {
		sm.startSync()
	}
}

// handleDonePeerMsg deals with peers that have signalled they are done.  It
// removes the peer as a candidate for syncing and in the case where it was
// the current sync peer, attempts to select a new best peer to sync from.  It
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

	// Attempt to find a new peer to sync from if the quitting peer is the
	// sync peer.  Also, reset the headers-first state if in headers-first
	// mode so
	if sm.syncPeer == peer {
		sm.syncPeer = nil
		if sm.headersFirstMode {
			selectedTipHash := sm.dag.SelectedTipHash()
			sm.resetHeaderState(selectedTipHash, sm.dag.ChainHeight()) //TODO: (Ori) This is probably wrong. Done only for compilation
		}
		sm.startSync()
	}
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

	// Ignore transactions that we have already rejected.  Do not
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

	// Remove transaction from request maps. Either the mempool/chain
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
		// so log it as such.  Otherwise, something really did go wrong,
		// so log it as an actual error.
		if _, ok := err.(mempool.RuleError); ok {
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
// We consider ourselves current iff at least one of the following is true:
// 1. there's no syncPeer, a.k.a. all connected peers are at the same tip
// 2. the DAG considers itself current - to prevent attacks where a peer sends an
//    unknown tip but never lets us sync to it.
func (sm *SyncManager) current() bool {
	return sm.syncPeer == nil || sm.dag.IsCurrent()
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
				return len(syncPeerState.requestedBlocks) != 0 || len(syncPeerState.requestQueue) != 0
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
		// to test duplicate block insertion fails.  Don't disconnect
		// the peer or ignore the block when we're in regression test
		// mode in this case so the chain code is actually fed the
		// duplicate blocks.
		if sm.chainParams != &dagconfig.RegressionNetParams {
			log.Warnf("Got unrequested block %s from %s -- "+
				"disconnecting", blockHash, peer.Addr())
			peer.Disconnect()
			return
		}
	}

	// When in headers-first mode, if the block matches the hash of the
	// first header in the list of headers that are being fetched, it's
	// eligible for less validation since the headers have already been
	// verified to link together and are valid up to the next checkpoint.
	// Also, remove the list entry for all blocks except the checkpoint
	// since it is needed to verify the next round of headers links
	// properly.
	isCheckpointBlock := false
	behaviorFlags := blockdag.BFNone
	if sm.headersFirstMode {
		firstNodeEl := sm.headerList.Front()
		if firstNodeEl != nil {
			firstNode := firstNodeEl.Value.(*headerNode)
			if blockHash.IsEqual(firstNode.hash) {
				behaviorFlags |= blockdag.BFFastAdd
				if firstNode.hash.IsEqual(sm.nextCheckpoint.Hash) {
					isCheckpointBlock = true
				} else {
					sm.headerList.Remove(firstNodeEl)
				}
			}
		}
	}

	if bmsg.isDelayedBlock {
		behaviorFlags |= blockdag.BFAfterDelay
	}

	// Process the block to include validation, orphan handling, etc.
	isOrphan, delay, err := sm.dag.ProcessBlock(bmsg.block, behaviorFlags)

	// Remove block from request maps. Either DAG knows about it and
	// so we shouldn't have any more instances of trying to fetch it, or
	// the insertion fails and thus we'll retry next time we get an inv.
	delete(state.requestedBlocks, *blockHash)
	delete(sm.requestedBlocks, *blockHash)

	sm.restartSyncIfNeeded()

	if err != nil {
		// When the error is a rule error, it means the block was simply
		// rejected as opposed to something actually going wrong, so log
		// it as such.  Otherwise, something really did go wrong, so log
		// it as an actual error.
		if _, ok := err.(blockdag.RuleError); ok {
			log.Infof("Rejected block %s from %s: %s", blockHash,
				peer, err)
		} else {
			log.Errorf("Failed to process block %s: %s",
				blockHash, err)
		}
		if dbErr, ok := err.(database.Error); ok && dbErr.ErrorCode ==
			database.ErrCorruption {
			panic(dbErr)
		}

		// Convert the error into an appropriate reject message and
		// send it.
		code, reason := mempool.ErrToRejectErr(err)
		peer.PushRejectMsg(wire.CmdBlock, code, reason, blockHash, false)
		return
	}

	if delay != 0 {
		spawn(func() {
			sm.QueueBlock(bmsg.block, bmsg.peer, true, make(chan struct{}))
		})
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

	// Nothing more to do if we aren't in headers-first mode.
	if !sm.headersFirstMode {
		return
	}

	// This is headers-first mode, so if the block is not a checkpoint
	// request more blocks using the header list when the request queue is
	// getting short.
	if !isCheckpointBlock {
		if sm.startHeader != nil &&
			len(state.requestedBlocks) < minInFlightBlocks {
			sm.fetchHeaderBlocks()
		}
		return
	}

	// This is headers-first mode and the block is a checkpoint.  When
	// there is a next checkpoint, get the next round of headers by asking
	// for headers starting from the block after this one up to the next
	// checkpoint.
	prevHeight := sm.nextCheckpoint.ChainHeight
	parentHash := sm.nextCheckpoint.Hash
	sm.nextCheckpoint = sm.findNextHeaderCheckpoint(prevHeight)
	if sm.nextCheckpoint != nil {
		locator := blockdag.BlockLocator([]*daghash.Hash{parentHash})
		err := peer.PushGetHeadersMsg(locator, sm.nextCheckpoint.Hash)
		if err != nil {
			log.Warnf("Failed to send getheaders message to "+
				"peer %s: %s", peer.Addr(), err)
			return
		}
		log.Infof("Downloading headers for blocks %d to %d from "+
			"peer %s", prevHeight+1, sm.nextCheckpoint.ChainHeight,
			sm.syncPeer.Addr())
		return
	}

	// This is headers-first mode, the block is a checkpoint, and there are
	// no more checkpoints, so switch to normal mode by requesting blocks
	// from the block after this one up to the end of the chain (zero hash).
	sm.headersFirstMode = false
	sm.headerList.Init()
	log.Infof("Reached the final checkpoint -- switching to normal mode")
	locator := blockdag.BlockLocator([]*daghash.Hash{blockHash})
	err = peer.PushGetBlocksMsg(locator, &daghash.ZeroHash)
	if err != nil {
		log.Warnf("Failed to send getblocks message to peer %s: %s",
			peer.Addr(), err)
		return
	}
}

func (sm *SyncManager) addBlocksToRequestQueue(state *peerSyncState, hashes []*daghash.Hash, isRelayedInv bool) {
	state.requestQueueMtx.Lock()
	defer state.requestQueueMtx.Unlock()
	for _, hash := range hashes {
		if _, exists := sm.requestedBlocks[*hash]; !exists {
			iv := wire.NewInvVect(wire.InvTypeBlock, hash)
			state.addInvToRequestQueueNoLock(iv, isRelayedInv)
		}
	}
}

func (state *peerSyncState) addInvToRequestQueueNoLock(iv *wire.InvVect, isRelayedInv bool) {
	if isRelayedInv {
		if _, exists := state.relayedInvsRequestQueueSet[*iv.Hash]; !exists {
			state.relayedInvsRequestQueueSet[*iv.Hash] = struct{}{}
			state.relayedInvsRequestQueue = append(state.relayedInvsRequestQueue, iv)
		}
	} else {
		if _, exists := state.requestQueueSet[*iv.Hash]; !exists {
			state.requestQueueSet[*iv.Hash] = struct{}{}
			state.requestQueue = append(state.requestQueue, iv)
		}
	}
}

func (state *peerSyncState) addInvToRequestQueue(iv *wire.InvVect, isRelayedInv bool) {
	state.requestQueueMtx.Lock()
	defer state.requestQueueMtx.Unlock()
	state.addInvToRequestQueueNoLock(iv, isRelayedInv)
}

// fetchHeaderBlocks creates and sends a request to the syncPeer for the next
// list of blocks to be downloaded based on the current list of headers.
func (sm *SyncManager) fetchHeaderBlocks() {
	// Nothing to do if there is no start header.
	if sm.startHeader == nil {
		log.Warnf("fetchHeaderBlocks called with no start header")
		return
	}

	// Build up a getdata request for the list of blocks the headers
	// describe.  The size hint will be limited to wire.MaxInvPerMsg by
	// the function, so no need to double check it here.
	gdmsg := wire.NewMsgGetDataSizeHint(uint(sm.headerList.Len()))
	numRequested := 0
	for e := sm.startHeader; e != nil; e = e.Next() {
		node, ok := e.Value.(*headerNode)
		if !ok {
			log.Warn("Header list node type is not a headerNode")
			continue
		}

		iv := wire.NewInvVect(wire.InvTypeBlock, node.hash)
		haveInv, err := sm.haveInventory(iv)
		if err != nil {
			log.Warnf("Unexpected failure when checking for "+
				"existing inventory during header block "+
				"fetch: %s", err)
		}
		if !haveInv {
			syncPeerState := sm.peerStates[sm.syncPeer]

			sm.requestedBlocks[*node.hash] = struct{}{}
			syncPeerState.requestedBlocks[*node.hash] = struct{}{}

			gdmsg.AddInvVect(iv)
			numRequested++
		}
		sm.startHeader = e.Next()
		if numRequested >= wire.MaxInvPerMsg {
			break
		}
	}
	if len(gdmsg.InvList) > 0 {
		sm.syncPeer.QueueMessage(gdmsg, nil)
	}
}

// handleHeadersMsg handles block header messages from all peers.  Headers are
// requested when performing a headers-first sync.
func (sm *SyncManager) handleHeadersMsg(hmsg *headersMsg) {
	peer := hmsg.peer
	_, exists := sm.peerStates[peer]
	if !exists {
		log.Warnf("Received headers message from unknown peer %s", peer)
		return
	}

	// The remote peer is misbehaving if we didn't request headers.
	msg := hmsg.headers
	numHeaders := len(msg.Headers)
	if !sm.headersFirstMode {
		log.Warnf("Got %d unrequested headers from %s -- "+
			"disconnecting", numHeaders, peer.Addr())
		peer.Disconnect()
		return
	}

	// Nothing to do for an empty headers message.
	if numHeaders == 0 {
		return
	}

	// Process all of the received headers ensuring each one connects to the
	// previous and that checkpoints match.
	receivedCheckpoint := false
	var finalHash *daghash.Hash
	for _, blockHeader := range msg.Headers {
		blockHash := blockHeader.BlockHash()
		finalHash = blockHash

		// Ensure there is a previous header to compare against.
		prevNodeEl := sm.headerList.Back()
		if prevNodeEl == nil {
			log.Warnf("Header list does not contain a previous" +
				"element as expected -- disconnecting peer")
			peer.Disconnect()
			return
		}

		// Ensure the header properly connects to the previous one and
		// add it to the list of headers.
		node := headerNode{hash: blockHash}
		prevNode := prevNodeEl.Value.(*headerNode)
		if prevNode.hash.IsEqual(blockHeader.ParentHashes[0]) { // TODO: (Stas) This is wrong. Modified only to satisfy compilation.
			node.height = prevNode.height + 1
			e := sm.headerList.PushBack(&node)
			if sm.startHeader == nil {
				sm.startHeader = e
			}
		} else {
			log.Warnf("Received block header that does not "+
				"properly connect to the chain from peer %s "+
				"-- disconnecting", peer.Addr())
			peer.Disconnect()
			return
		}

		// Verify the header at the next checkpoint height matches.
		if node.height == sm.nextCheckpoint.ChainHeight {
			if node.hash.IsEqual(sm.nextCheckpoint.Hash) {
				receivedCheckpoint = true
				log.Infof("Verified downloaded block "+
					"header against checkpoint at height "+
					"%d/hash %s", node.height, node.hash)
			} else {
				log.Warnf("Block header at height %d/hash "+
					"%s from peer %s does NOT match "+
					"expected checkpoint hash of %s -- "+
					"disconnecting", node.height,
					node.hash, peer.Addr(),
					sm.nextCheckpoint.Hash)
				peer.Disconnect()
				return
			}
			break
		}
	}

	// When this header is a checkpoint, switch to fetching the blocks for
	// all of the headers since the last checkpoint.
	if receivedCheckpoint {
		// Since the first entry of the list is always the final block
		// that is already in the database and is only used to ensure
		// the next header links properly, it must be removed before
		// fetching the blocks.
		sm.headerList.Remove(sm.headerList.Front())
		log.Infof("Received %d block headers: Fetching blocks",
			sm.headerList.Len())
		sm.progressLogger.SetLastLogTime(time.Now())
		sm.fetchHeaderBlocks()
		return
	}

	// This header is not a checkpoint, so request the next batch of
	// headers starting from the latest known header and ending with the
	// next checkpoint.
	locator := blockdag.BlockLocator([]*daghash.Hash{finalHash})
	err := peer.PushGetHeadersMsg(locator, sm.nextCheckpoint.Hash)
	if err != nil {
		log.Warnf("Failed to send getheaders message to "+
			"peer %s: %s", peer.Addr(), err)
		return
	}
}

// haveInventory returns whether or not the inventory represented by the passed
// inventory vector is known.  This includes checking all of the various places
// inventory can be when it is in different states such as blocks that are part
// of the main chain, on a side chain, in the orphan pool, and transactions that
// are in the memory pool (either the main pool or orphan pool).
func (sm *SyncManager) haveInventory(invVect *wire.InvVect) (bool, error) {
	switch invVect.Type {
	case wire.InvTypeSyncBlock:
		fallthrough
	case wire.InvTypeBlock:
		// Ask DAG if the block is known to it in any form (in DAG or as an orphan).
		return sm.dag.HaveBlock(invVect.Hash)

	case wire.InvTypeTx:
		// Ask the transaction memory pool if the transaction is known
		// to it in any form (main pool or orphan).
		if sm.txMemPool.HaveTransaction((*daghash.TxID)(invVect.Hash)) {
			return true, nil
		}

		// Check if the transaction exists from the point of view of the
		// end of the main chain.  Note that this is only a best effort
		// since it is expensive to check existence of every output and
		// the only purpose of this check is to avoid downloading
		// already known transactions.  Only the first two outputs are
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

	// Attempt to find the final block in the inventory list.  There may
	// not be one.
	lastBlock := -1
	invVects := imsg.inv.InvList
	for i := len(invVects) - 1; i >= 0; i-- {
		if invVects[i].IsBlockOrSyncBlock() {
			lastBlock = i
			break
		}
	}

	// Request the advertised inventory if we don't already have it.  Also,
	// request parent blocks of orphans if we receive one we already have.
	// Finally, attempt to detect potential stalls due to long side chains
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

		// Ignore inventory when we're in headers-first mode.
		if sm.headersFirstMode {
			continue
		}

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
				// Skip the transaction if it has already been
				// rejected.
				if _, exists := sm.rejectedTxns[daghash.TxID(*iv.Hash)]; exists {
					continue
				}
			}

			// Add it to the request queue.
			state.addInvToRequestQueue(iv, iv.Type != wire.InvTypeSyncBlock)
			continue
		}

		if iv.IsBlockOrSyncBlock() {
			// The block is an orphan block that we already have.
			// When the existing orphan was processed, it requested
			// the missing parent blocks.  When this scenario
			// happens, it means there were more blocks missing
			// than are allowed into a single inventory message.  As
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
			// inventory message, so force a request for more.  This
			// should only happen if our DAG and the peer's DAG have
			// diverged long time ago.
			if i == lastBlock && peer == sm.syncPeer {
				// Request blocks after the first block's ancestor that exists
				// in the selected path chain, one up to the
				// final one the remote peer knows about (zero
				// stop hash).
				locator := sm.dag.BlockLocatorFromHash(iv.Hash)
				peer.PushGetBlocksMsg(locator, &daghash.ZeroHash)
			}
		}
	}

	err := sm.sendInvsFromRequestQueue(peer, state)
	if err != nil {
		log.Errorf("Failed to send invs from queue: %s", err)
	}
}

func (sm *SyncManager) addInvsToGetDataMessageFromQueue(gdmsg *wire.MsgGetData, state *peerSyncState, requestQueue []*wire.InvVect) ([]*wire.InvVect, error) {
	var invsNum int
	leftSpaceInGdmsg := wire.MaxInvPerMsg - len(gdmsg.InvList)
	if len(requestQueue) > leftSpaceInGdmsg {
		invsNum = leftSpaceInGdmsg
	} else {
		invsNum = len(requestQueue)
	}
	invsToAdd := make([]*wire.InvVect, 0, invsNum)

	for len(requestQueue) != 0 {
		iv := requestQueue[0]
		requestQueue[0] = nil
		requestQueue = requestQueue[1:]

		exists, err := sm.haveInventory(iv)
		if err != nil {
			return nil, err
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
		switch iv.Type {
		case wire.InvTypeSyncBlock:
			delete(state.requestQueueSet, *iv.Hash)
			addBlockInv(iv)
		case wire.InvTypeBlock:
			delete(state.relayedInvsRequestQueueSet, *iv.Hash)
			addBlockInv(iv)

		case wire.InvTypeTx:
			delete(state.relayedInvsRequestQueueSet, *iv.Hash)
			// Request the transaction if there is not already a
			// pending request.
			if _, exists := sm.requestedTxns[daghash.TxID(*iv.Hash)]; !exists {
				sm.requestedTxns[daghash.TxID(*iv.Hash)] = struct{}{}
				sm.limitTxIDMap(sm.requestedTxns, maxRequestedTxns)
				state.requestedTxns[daghash.TxID(*iv.Hash)] = struct{}{}

				gdmsg.AddInvVect(iv)
			}
		}

		if len(requestQueue) >= wire.MaxInvPerMsg {
			break
		}
	}
	return requestQueue, nil
}

func (sm *SyncManager) sendInvsFromRequestQueue(peer *peerpkg.Peer, state *peerSyncState) error {
	state.requestQueueMtx.Lock()
	defer state.requestQueueMtx.Unlock()
	gdmsg := wire.NewMsgGetData()
	newRequestQueue, err := sm.addInvsToGetDataMessageFromQueue(gdmsg, state, state.requestQueue)
	if err != nil {
		return err
	}
	state.requestQueue = newRequestQueue
	if sm.current() {
		newRequestQueue, err := sm.addInvsToGetDataMessageFromQueue(gdmsg, state, state.relayedInvsRequestQueue)
		if err != nil {
			return err
		}
		state.relayedInvsRequestQueue = newRequestQueue
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
		// Remove a random entry from the map.  For most compilers, Go's
		// range statement iterates starting at a random item although
		// that is not 100% guaranteed by the spec.  The iteration order
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
		// Remove a random entry from the map.  For most compilers, Go's
		// range statement iterates starting at a random item although
		// that is not 100% guaranteed by the spec.  The iteration order
		// is not important here because an adversary would have to be
		// able to pull off preimage attacks on the hashing function in
		// order to target eviction of specific entries anyways.
		for hash := range m {
			delete(m, hash)
			return
		}
	}
}

// blockHandler is the main handler for the sync manager.  It must be run as a
// goroutine.  It processes block and inv messages in a separate goroutine
// from the peer handlers so the block (MsgBlock) messages are handled by a
// single thread without needing to lock memory data structures.  This is
// important because the sync manager controls which blocks are needed and how
// the fetching should proceed.
func (sm *SyncManager) blockHandler() {
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

			case *headersMsg:
				sm.handleHeadersMsg(msg)

			case *donePeerMsg:
				sm.handleDonePeerMsg(msg.peer)

			case getSyncPeerMsg:
				var peerID int32
				if sm.syncPeer != nil {
					peerID = sm.syncPeer.ID()
				}
				msg.reply <- peerID

			case processBlockMsg:
				isOrphan, delay, err := sm.dag.ProcessBlock(
					msg.block, msg.flags)
				if err != nil {
					msg.reply <- processBlockResponse{
						isOrphan: false,
						err:      err,
					}
				}
				if delay != 0 {
					msg.reply <- processBlockResponse{
						isOrphan: false,
						err:      errors.New("Cannot process blocks from RPC beyond the allowed time offset"),
					}
				}

				msg.reply <- processBlockResponse{
					isOrphan: isOrphan,
					err:      nil,
				}

			case isCurrentMsg:
				msg.reply <- sm.current()

			case pauseMsg:
				// Wait until the sender unpauses the manager.
				<-msg.unpause

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

// handleBlockDAGNotification handles notifications from blockDAG.  It does
// things such as request orphan block parents and relay accepted blocks to
// connected peers.
func (sm *SyncManager) handleBlockDAGNotification(notification *blockdag.Notification) {
	switch notification.Type {
	// A block has been accepted into the blockDAG.  Relay it to other peers.
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

		// Don't relay if we are not current or the block was just now unorphaned.
		// Other peers that are current should already know about it
		if !sm.current() || data.WasUnorphaned {
			return
		}

		// Generate the inventory vector and relay it.
		iv := wire.NewInvVect(wire.InvTypeBlock, block.Hash())
		sm.peerNotifier.RelayInventory(iv, block.MsgBlock().Header)

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

// QueueHeaders adds the passed headers message and peer to the block handling
// queue.
func (sm *SyncManager) QueueHeaders(headers *wire.MsgHeaders, peer *peerpkg.Peer) {
	// No channel handling here because peers do not need to block on
	// headers messages.
	if atomic.LoadInt32(&sm.shutdown) != 0 {
		return
	}

	sm.msgChan <- &headersMsg{headers: headers, peer: peer}
}

// DonePeer informs the blockmanager that a peer has disconnected.
func (sm *SyncManager) DonePeer(peer *peerpkg.Peer) {
	// Ignore if we are shutting down.
	if atomic.LoadInt32(&sm.shutdown) != 0 {
		return
	}

	sm.msgChan <- &donePeerMsg{peer: peer}
}

// Start begins the core block handler which processes block and inv messages.
func (sm *SyncManager) Start() {
	// Already started?
	if atomic.AddInt32(&sm.started, 1) != 1 {
		return
	}

	log.Trace("Starting sync manager")
	sm.wg.Add(1)
	spawn(sm.blockHandler)
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
	reply := make(chan processBlockResponse, 1)
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
// Note that while paused, all peer and block processing is halted.  The
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
		chainParams:     config.ChainParams,
		rejectedTxns:    make(map[daghash.TxID]struct{}),
		requestedTxns:   make(map[daghash.TxID]struct{}),
		requestedBlocks: make(map[daghash.Hash]struct{}),
		peerStates:      make(map[*peerpkg.Peer]*peerSyncState),
		progressLogger:  newBlockProgressLogger("Processed", log),
		msgChan:         make(chan interface{}, config.MaxPeers*3),
		headerList:      list.New(),
		quit:            make(chan struct{}),
	}

	selectedTipHash := sm.dag.SelectedTipHash()
	if !config.DisableCheckpoints {
		// Initialize the next checkpoint based on the current chain height.
		sm.nextCheckpoint = sm.findNextHeaderCheckpoint(sm.dag.ChainHeight()) //TODO: (Ori) This is probably wrong. Done only for compilation
		if sm.nextCheckpoint != nil {
			sm.resetHeaderState(selectedTipHash, sm.dag.ChainHeight()) //TODO: (Ori) This is probably wrong. Done only for compilation)
		}
	} else {
		log.Info("Checkpoints are disabled")
	}

	sm.dag.Subscribe(sm.handleBlockDAGNotification)

	return &sm, nil
}
