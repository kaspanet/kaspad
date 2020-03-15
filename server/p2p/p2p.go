// Copyright (c) 2013-2017 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package p2p

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/util/subnetworkid"

	"github.com/kaspanet/kaspad/addrmgr"
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/blockdag/indexers"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/connmgr"
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/database"
	"github.com/kaspanet/kaspad/logger"
	"github.com/kaspanet/kaspad/mempool"
	"github.com/kaspanet/kaspad/netsync"
	"github.com/kaspanet/kaspad/peer"
	"github.com/kaspanet/kaspad/server/serverutils"
	"github.com/kaspanet/kaspad/txscript"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/bloom"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/version"
	"github.com/kaspanet/kaspad/wire"
)

const (
	// defaultServices describes the default services that are supported by
	// the server.
	defaultServices = wire.SFNodeNetwork | wire.SFNodeBloom | wire.SFNodeCF

	// defaultRequiredServices describes the default services that are
	// required to be supported by outbound peers.
	defaultRequiredServices = wire.SFNodeNetwork

	// connectionRetryInterval is the base amount of time to wait in between
	// retries when connecting to persistent peers. It is adjusted by the
	// number of retries such that there is a retry backoff.
	connectionRetryInterval = time.Second * 5
)

var (
	// userAgentName is the user agent name and is used to help identify
	// ourselves to other kaspa peers.
	userAgentName = "kaspad"

	// userAgentVersion is the user agent version and is used to help
	// identify ourselves to other kaspa peers.
	userAgentVersion = version.Version()
)

// simpleAddr implements the net.Addr interface with two struct fields
type simpleAddr struct {
	net, addr string
}

// String returns the address.
//
// This is part of the net.Addr interface.
func (a simpleAddr) String() string {
	return a.addr
}

// Network returns the network.
//
// This is part of the net.Addr interface.
func (a simpleAddr) Network() string {
	return a.net
}

// Ensure simpleAddr implements the net.Addr interface.
var _ net.Addr = simpleAddr{}

// broadcastMsg provides the ability to house a kaspa message to be broadcast
// to all connected peers except specified excluded peers.
type broadcastMsg struct {
	message      wire.Message
	excludePeers []*Peer
}

// broadcastInventoryAdd is a type used to declare that the InvVect it contains
// needs to be added to the rebroadcast map
type broadcastInventoryAdd relayMsg

// broadcastInventoryDel is a type used to declare that the InvVect it contains
// needs to be removed from the rebroadcast map
type broadcastInventoryDel *wire.InvVect

// relayMsg packages an inventory vector along with the newly discovered
// inventory so the relay has access to that information.
type relayMsg struct {
	invVect *wire.InvVect
	data    interface{}
}

type outboundPeerConnectedMsg struct {
	connReq *connmgr.ConnReq
	conn    net.Conn
}

// Peer extends the peer to maintain state shared by the server and
// the blockmanager.
type Peer struct {
	// The following variables must only be used atomically
	FeeFilterInt int64

	*peer.Peer

	connReq         *connmgr.ConnReq
	server          *Server
	persistent      bool
	relayMtx        sync.Mutex
	DisableRelayTx  bool
	sentAddrs       bool
	isWhitelisted   bool
	filter          *bloom.Filter
	knownAddresses  map[string]struct{}
	DynamicBanScore connmgr.DynamicBanScore
	quit            chan struct{}
	// The following chans are used to sync blockmanager and server.
	txProcessed    chan struct{}
	blockProcessed chan struct{}
}

// peerState maintains state of inbound, persistent, outbound peers as well
// as banned peers and outbound groups.
type peerState struct {
	inboundPeers    map[int32]*Peer
	outboundPeers   map[int32]*Peer
	persistentPeers map[int32]*Peer
	banned          map[string]time.Time
	outboundGroups  map[string]int
}

// Count returns the count of all known peers.
func (ps *peerState) Count() int {
	return ps.countInboundPeers() + ps.countOutboundPeers()
}

func (ps *peerState) countInboundPeers() int {
	return len(ps.inboundPeers)
}

func (ps *peerState) countOutboundPeers() int {
	return len(ps.outboundPeers) +
		len(ps.persistentPeers)
}

// forAllOutboundPeers is a helper function that runs a callback on all outbound
// peers known to peerState.
// The loop stops and returns false if one of the callback calls returns false.
// Otherwise the function should return true.
func (ps *peerState) forAllOutboundPeers(callback func(sp *Peer) bool) bool {
	for _, e := range ps.outboundPeers {
		shouldContinue := callback(e)
		if !shouldContinue {
			return false
		}
	}
	for _, e := range ps.persistentPeers {
		shouldContinue := callback(e)
		if !shouldContinue {
			return false
		}
	}
	return true
}

// forAllInboundPeers is a helper function that runs a callback on all inbound
// peers known to peerState.
// The loop stops and returns false if one of the callback calls returns false.
// Otherwise the function should return true.
func (ps *peerState) forAllInboundPeers(callback func(sp *Peer) bool) bool {
	for _, e := range ps.inboundPeers {
		shouldContinue := callback(e)
		if !shouldContinue {
			return false
		}
	}
	return true
}

// forAllPeers is a helper function that runs a callback on all peers known to
// peerState.
// The loop stops and returns false if one of the callback calls returns false.
// Otherwise the function should return true.
func (ps *peerState) forAllPeers(callback func(sp *Peer) bool) bool {
	shouldContinue := ps.forAllInboundPeers(callback)
	if !shouldContinue {
		return false
	}
	return ps.forAllOutboundPeers(callback)
}

// Server provides a kaspa server for handling communications to and from
// kaspa peers.
type Server struct {
	// The following variables must only be used atomically.
	// Putting the uint64s first makes them 64-bit aligned for 32-bit systems.
	bytesReceived uint64 // Total bytes received from all peers since start.
	bytesSent     uint64 // Total bytes sent by all peers since start.
	shutdown      int32
	shutdownSched int32

	DAGParams   *dagconfig.Params
	addrManager *addrmgr.AddrManager
	connManager *connmgr.ConnManager
	SigCache    *txscript.SigCache
	SyncManager *netsync.SyncManager
	DAG         *blockdag.BlockDAG
	TxMemPool   *mempool.TxPool

	modifyRebroadcastInv  chan interface{}
	newPeers              chan *Peer
	donePeers             chan *Peer
	banPeers              chan *Peer
	newOutboundConnection chan *outboundPeerConnectedMsg
	Query                 chan interface{}
	relayInv              chan relayMsg
	broadcast             chan broadcastMsg
	wg                    sync.WaitGroup
	nat                   serverutils.NAT
	db                    database.DB
	TimeSource            blockdag.TimeSource
	services              wire.ServiceFlag

	// We add to quitWaitGroup before every instance in which we wait for
	// the quit channel so that all those instances finish before we shut
	// down the managers (connManager, addrManager, etc),
	quitWaitGroup sync.WaitGroup
	quit          chan struct{}

	// The following fields are used for optional indexes. They will be nil
	// if the associated index is not enabled. These fields are set during
	// initial creation of the server and never changed afterwards, so they
	// do not need to be protected for concurrent access.
	AcceptanceIndex *indexers.AcceptanceIndex

	notifyNewTransactions func(txns []*mempool.TxDesc)
}

// newServerPeer returns a new serverPeer instance. The peer needs to be set by
// the caller.
func newServerPeer(s *Server, isPersistent bool) *Peer {
	return &Peer{
		server:         s,
		persistent:     isPersistent,
		filter:         bloom.LoadFilter(nil),
		knownAddresses: make(map[string]struct{}),
		quit:           make(chan struct{}),
		txProcessed:    make(chan struct{}, 1),
		blockProcessed: make(chan struct{}, 1),
	}
}

// selectedTipHash returns the current selected tip hash
func (sp *Peer) selectedTipHash() *daghash.Hash {
	return sp.server.DAG.SelectedTipHash()
}

// blockExists determines whether a block with the given hash exists in
// the DAG.
func (sp *Peer) blockExists(hash *daghash.Hash) bool {
	return sp.server.DAG.IsInDAG(hash)
}

// addKnownAddresses adds the given addresses to the set of known addresses to
// the peer to prevent sending duplicate addresses.
func (sp *Peer) addKnownAddresses(addresses []*wire.NetAddress) {
	for _, na := range addresses {
		sp.knownAddresses[addrmgr.NetAddressKey(na)] = struct{}{}
	}
}

// addressKnown true if the given address is already known to the peer.
func (sp *Peer) addressKnown(na *wire.NetAddress) bool {
	_, exists := sp.knownAddresses[addrmgr.NetAddressKey(na)]
	return exists
}

// setDisableRelayTx toggles relaying of transactions for the given peer.
// It is safe for concurrent access.
func (sp *Peer) setDisableRelayTx(disable bool) {
	sp.relayMtx.Lock()
	defer sp.relayMtx.Unlock()
	sp.DisableRelayTx = disable
}

// relayTxDisabled returns whether or not relaying of transactions for the given
// peer is disabled.
// It is safe for concurrent access.
func (sp *Peer) relayTxDisabled() bool {
	sp.relayMtx.Lock()
	defer sp.relayMtx.Unlock()
	return sp.DisableRelayTx
}

// pushAddrMsg sends an addr message to the connected peer using the provided
// addresses.
func (sp *Peer) pushAddrMsg(addresses []*wire.NetAddress, subnetworkID *subnetworkid.SubnetworkID) {
	// Filter addresses already known to the peer.
	addrs := make([]*wire.NetAddress, 0, len(addresses))
	for _, addr := range addresses {
		if !sp.addressKnown(addr) {
			addrs = append(addrs, addr)
		}
	}
	known, err := sp.PushAddrMsg(addrs, subnetworkID)
	if err != nil {
		peerLog.Errorf("Can't push address message to %s: %s", sp.Peer, err)
		sp.Disconnect()
		return
	}
	sp.addKnownAddresses(known)
}

// addBanScore increases the persistent and decaying ban score fields by the
// values passed as parameters. If the resulting score exceeds half of the ban
// threshold, a warning is logged including the reason provided. Further, if
// the score is above the ban threshold, the peer will be banned and
// disconnected.
func (sp *Peer) addBanScore(persistent, transient uint32, reason string) {
	// No warning is logged and no score is calculated if banning is disabled.
	if config.ActiveConfig().DisableBanning {
		return
	}
	if sp.isWhitelisted {
		peerLog.Debugf("Misbehaving whitelisted peer %s: %s", sp, reason)
		return
	}

	warnThreshold := config.ActiveConfig().BanThreshold >> 1
	if transient == 0 && persistent == 0 {
		// The score is not being increased, but a warning message is still
		// logged if the score is above the warn threshold.
		score := sp.DynamicBanScore.Int()
		if score > warnThreshold {
			peerLog.Warnf("Misbehaving peer %s: %s -- ban score is %d, "+
				"it was not increased this time", sp, reason, score)
		}
		return
	}
	score := sp.DynamicBanScore.Increase(persistent, transient)
	if score > warnThreshold {
		peerLog.Warnf("Misbehaving peer %s: %s -- ban score increased to %d",
			sp, reason, score)
		if score > config.ActiveConfig().BanThreshold {
			peerLog.Warnf("Misbehaving peer %s -- banning and disconnecting",
				sp)
			sp.server.BanPeer(sp)
			sp.Disconnect()
		}
	}
}

// enforceNodeBloomFlag disconnects the peer if the server is not configured to
// allow bloom filters. Additionally, if the peer has negotiated to a protocol
// version  that is high enough to observe the bloom filter service support bit,
// it will be banned since it is intentionally violating the protocol.
func (sp *Peer) enforceNodeBloomFlag(cmd string) bool {
	if sp.server.services&wire.SFNodeBloom != wire.SFNodeBloom {
		// NOTE: Even though the addBanScore function already examines
		// whether or not banning is enabled, it is checked here as well
		// to ensure the violation is logged and the peer is
		// disconnected regardless.
		if !config.ActiveConfig().DisableBanning {

			// Disconnect the peer regardless of whether it was
			// banned.
			sp.addBanScore(100, 0, cmd)
			sp.Disconnect()
			return false
		}

		// Disconnect the peer regardless of protocol version or banning
		// state.
		peerLog.Debugf("%s sent an unsupported %s request -- "+
			"disconnecting", sp, cmd)
		sp.Disconnect()
		return false
	}

	return true
}

// OnRead is invoked when a peer receives a message and it is used to update
// the bytes received by the server.
func (sp *Peer) OnRead(_ *peer.Peer, bytesRead int, msg wire.Message, err error) {
	sp.server.AddBytesReceived(uint64(bytesRead))
}

// OnWrite is invoked when a peer sends a message and it is used to update
// the bytes sent by the server.
func (sp *Peer) OnWrite(_ *peer.Peer, bytesWritten int, msg wire.Message, err error) {
	sp.server.AddBytesSent(uint64(bytesWritten))
}

// randomUint16Number returns a random uint16 in a specified input range. Note
// that the range is in zeroth ordering; if you pass it 1800, you will get
// values from 0 to 1800.
func randomUint16Number(max uint16) uint16 {
	// In order to avoid modulo bias and ensure every possible outcome in
	// [0, max) has equal probability, the random number must be sampled
	// from a random source that has a range limited to a multiple of the
	// modulus.
	var randomNumber uint16
	var limitRange = (math.MaxUint16 / max) * max
	for {
		binary.Read(rand.Reader, binary.LittleEndian, &randomNumber)
		if randomNumber < limitRange {
			return (randomNumber % max)
		}
	}
}

// AddRebroadcastInventory adds 'iv' to the list of inventories to be
// rebroadcasted at random intervals until they show up in a block.
func (s *Server) AddRebroadcastInventory(iv *wire.InvVect, data interface{}) {
	// Ignore if shutting down.
	if atomic.LoadInt32(&s.shutdown) != 0 {
		return
	}

	s.modifyRebroadcastInv <- broadcastInventoryAdd{invVect: iv, data: data}
}

// RemoveRebroadcastInventory removes 'iv' from the list of items to be
// rebroadcasted if present.
func (s *Server) RemoveRebroadcastInventory(iv *wire.InvVect) {
	// Ignore if shutting down.
	if atomic.LoadInt32(&s.shutdown) != 0 {
		return
	}

	s.modifyRebroadcastInv <- broadcastInventoryDel(iv)
}

// RelayTransactions generates and relays inventory vectors for all of the
// passed transactions to all connected peers.
func (s *Server) RelayTransactions(txns []*mempool.TxDesc) {
	for _, txD := range txns {
		iv := wire.NewInvVect(wire.InvTypeTx, (*daghash.Hash)(txD.Tx.ID()))
		s.RelayInventory(iv, txD)
	}
}

// pushTxMsg sends a tx message for the provided transaction hash to the
// connected peer. An error is returned if the transaction hash is not known.
func (s *Server) pushTxMsg(sp *Peer, txID *daghash.TxID, doneChan chan<- struct{},
	waitChan <-chan struct{}) error {

	// Attempt to fetch the requested transaction from the pool. A
	// call could be made to check for existence first, but simply trying
	// to fetch a missing transaction results in the same behavior.
	tx, err := s.TxMemPool.FetchTransaction(txID)
	if err != nil {
		peerLog.Tracef("Unable to fetch tx %s from transaction "+
			"pool: %s", txID, err)

		if doneChan != nil {
			doneChan <- struct{}{}
		}
		return err
	}

	// Once we have fetched data wait for any previous operation to finish.
	if waitChan != nil {
		<-waitChan
	}

	sp.QueueMessage(tx.MsgTx(), doneChan)

	return nil
}

// pushBlockMsg sends a block message for the provided block hash to the
// connected peer. An error is returned if the block hash is not known.
func (s *Server) pushBlockMsg(sp *Peer, hash *daghash.Hash, doneChan chan<- struct{},
	waitChan <-chan struct{}) error {

	// Fetch the raw block bytes from the database.
	var blockBytes []byte
	err := sp.server.db.View(func(dbTx database.Tx) error {
		var err error
		blockBytes, err = dbTx.FetchBlock(hash)
		return err
	})
	if err != nil {
		peerLog.Tracef("Unable to fetch requested block hash %s: %s",
			hash, err)

		if doneChan != nil {
			doneChan <- struct{}{}
		}
		return err
	}

	// Deserialize the block.
	var msgBlock wire.MsgBlock
	err = msgBlock.Deserialize(bytes.NewReader(blockBytes))
	if err != nil {
		peerLog.Tracef("Unable to deserialize requested block hash "+
			"%s: %s", hash, err)

		if doneChan != nil {
			doneChan <- struct{}{}
		}
		return err
	}

	// If we are a full node and the peer is a partial node, we must convert
	// the block to a partial block.
	nodeSubnetworkID := s.DAG.SubnetworkID()
	peerSubnetworkID := sp.Peer.SubnetworkID()
	isNodeFull := nodeSubnetworkID == nil
	isPeerFull := peerSubnetworkID == nil
	if isNodeFull && !isPeerFull {
		msgBlock.ConvertToPartial(peerSubnetworkID)
	}

	// Once we have fetched data wait for any previous operation to finish.
	if waitChan != nil {
		<-waitChan
	}

	sp.QueueMessage(&msgBlock, doneChan)

	return nil
}

// pushMerkleBlockMsg sends a merkleblock message for the provided block hash to
// the connected peer. Since a merkle block requires the peer to have a filter
// loaded, this call will simply be ignored if there is no filter loaded. An
// error is returned if the block hash is not known.
func (s *Server) pushMerkleBlockMsg(sp *Peer, hash *daghash.Hash,
	doneChan chan<- struct{}, waitChan <-chan struct{}) error {

	// Do not send a response if the peer doesn't have a filter loaded.
	if !sp.filter.IsLoaded() {
		if doneChan != nil {
			doneChan <- struct{}{}
		}
		return nil
	}

	// Fetch the raw block bytes from the database.
	blk, err := sp.server.DAG.BlockByHash(hash)
	if err != nil {
		peerLog.Tracef("Unable to fetch requested block hash %s: %s",
			hash, err)

		if doneChan != nil {
			doneChan <- struct{}{}
		}
		return err
	}

	// Generate a merkle block by filtering the requested block according
	// to the filter for the peer.
	merkle, matchedTxIndices := bloom.NewMerkleBlock(blk, sp.filter)

	// Once we have fetched data wait for any previous operation to finish.
	if waitChan != nil {
		<-waitChan
	}

	// Send the merkleblock. Only send the done channel with this message
	// if no transactions will be sent afterwards.
	var dc chan<- struct{}
	if len(matchedTxIndices) == 0 {
		dc = doneChan
	}
	sp.QueueMessage(merkle, dc)

	// Finally, send any matched transactions.
	blkTransactions := blk.MsgBlock().Transactions
	for i, txIndex := range matchedTxIndices {
		// Only send the done channel on the final transaction.
		var dc chan<- struct{}
		if i == len(matchedTxIndices)-1 {
			dc = doneChan
		}
		if txIndex < uint32(len(blkTransactions)) {
			sp.QueueMessage(blkTransactions[txIndex], dc)
		}
	}

	return nil
}

// handleAddPeerMsg deals with adding new peers. It is invoked from the
// peerHandler goroutine.
func (s *Server) handleAddPeerMsg(state *peerState, sp *Peer) bool {
	if sp == nil {
		return false
	}

	// Ignore new peers if we're shutting down.
	if atomic.LoadInt32(&s.shutdown) != 0 {
		srvrLog.Infof("New peer %s ignored - server is shutting down", sp)
		sp.Disconnect()
		return false
	}

	// Disconnect banned peers.
	host, _, err := net.SplitHostPort(sp.Addr())
	if err != nil {
		srvrLog.Debugf("can't split hostport %s", err)
		sp.Disconnect()
		return false
	}
	if banEnd, ok := state.banned[host]; ok {
		if time.Now().Before(banEnd) {
			srvrLog.Debugf("Peer %s is banned for another %s - disconnecting",
				host, time.Until(banEnd))
			sp.Disconnect()
			return false
		}

		srvrLog.Infof("Peer %s is no longer banned", host)
		delete(state.banned, host)
	}

	// TODO: Check for max peers from a single IP.

	// Limit max number of total peers.
	if sp.Inbound() && len(state.inboundPeers) >= config.ActiveConfig().MaxInboundPeers {
		srvrLog.Infof("Max inbound peers reached [%d] - disconnecting peer %s",
			config.ActiveConfig().MaxInboundPeers, sp)
		sp.Disconnect()
		return false
	}

	// Add the new peer and start it.
	srvrLog.Debugf("New peer %s", sp)
	if sp.Inbound() {
		state.inboundPeers[sp.ID()] = sp
	} else {
		if sp.persistent {
			state.persistentPeers[sp.ID()] = sp
		} else {
			state.outboundPeers[sp.ID()] = sp
		}
	}

	// Notify the connection manager.
	s.connManager.NotifyConnectionRequestComplete()

	return true
}

// handleDonePeerMsg deals with peers that have signalled they are done. It is
// invoked from the peerHandler goroutine.
func (s *Server) handleDonePeerMsg(state *peerState, sp *Peer) {
	var list map[int32]*Peer
	if sp.persistent {
		list = state.persistentPeers
	} else if sp.Inbound() {
		list = state.inboundPeers
	} else {
		list = state.outboundPeers
	}
	if _, ok := list[sp.ID()]; ok {
		if !sp.Inbound() && sp.VersionKnown() {
			state.outboundGroups[addrmgr.GroupKey(sp.NA())]--
		}
		if !sp.Inbound() && sp.connReq != nil {
			s.connManager.Disconnect(sp.connReq.ID())
		}
		delete(list, sp.ID())
		srvrLog.Debugf("Removed peer %s", sp)
		return
	}

	if sp.connReq != nil {
		s.connManager.Disconnect(sp.connReq.ID())
	}

	// Update the address' last seen time if the peer has acknowledged
	// our version and has sent us its version as well.
	if sp.VerAckReceived() && sp.VersionKnown() && sp.NA() != nil {
		s.addrManager.Connected(sp.NA())
	}

	// If we get here it means that either we didn't know about the peer
	// or we purposefully deleted it.
}

// handleBanPeerMsg deals with banning peers. It is invoked from the
// peerHandler goroutine.
func (s *Server) handleBanPeerMsg(state *peerState, sp *Peer) {
	host, _, err := net.SplitHostPort(sp.Addr())
	if err != nil {
		srvrLog.Debugf("can't split ban peer %s: %s", sp.Addr(), err)
		return
	}
	direction := logger.DirectionString(sp.Inbound())
	srvrLog.Infof("Banned peer %s (%s) for %s", host, direction,
		config.ActiveConfig().BanDuration)
	state.banned[host] = time.Now().Add(config.ActiveConfig().BanDuration)
}

// handleRelayInvMsg deals with relaying inventory to peers that are not already
// known to have it. It is invoked from the peerHandler goroutine.
func (s *Server) handleRelayInvMsg(state *peerState, msg relayMsg) {
	state.forAllPeers(func(sp *Peer) bool {
		if !sp.Connected() {
			return true
		}

		if msg.invVect.Type == wire.InvTypeTx {
			// Don't relay the transaction to the peer when it has
			// transaction relaying disabled.
			if sp.relayTxDisabled() {
				return true
			}

			txD, ok := msg.data.(*mempool.TxDesc)
			if !ok {
				peerLog.Warnf("Underlying data for tx inv "+
					"relay is not a *mempool.TxDesc: %T",
					msg.data)
				return true
			}

			// Don't relay the transaction if the transaction fee-per-kb
			// is less than the peer's feefilter.
			feeFilter := uint64(atomic.LoadInt64(&sp.FeeFilterInt))
			if feeFilter > 0 && txD.FeePerKB < feeFilter {
				return true
			}

			// Don't relay the transaction if there is a bloom
			// filter loaded and the transaction doesn't match it.
			if sp.filter.IsLoaded() {
				if !sp.filter.MatchTxAndUpdate(txD.Tx) {
					return true
				}
			}

			// Don't relay the transaction if the peer's subnetwork is
			// incompatible with it.
			if !txD.Tx.MsgTx().IsSubnetworkCompatible(sp.Peer.SubnetworkID()) {
				return true
			}
		}

		// Queue the inventory to be relayed with the next batch.
		// It will be ignored if the peer is already known to
		// have the inventory.
		sp.QueueInventory(msg.invVect)
		return true
	})
}

// handleBroadcastMsg deals with broadcasting messages to peers. It is invoked
// from the peerHandler goroutine.
func (s *Server) handleBroadcastMsg(state *peerState, bmsg *broadcastMsg) {
	state.forAllPeers(func(sp *Peer) bool {
		if !sp.Connected() {
			return true
		}

		for _, ep := range bmsg.excludePeers {
			if sp == ep {
				return true
			}
		}

		sp.QueueMessage(bmsg.message, nil)
		return true
	})
}

type getConnCountMsg struct {
	reply chan int32
}

type getShouldMineOnGenesisMsg struct {
	reply chan bool
}

//GetPeersMsg is the message type which is used by the rpc server to get the peers list from the p2p server
type GetPeersMsg struct {
	Reply chan []*Peer
}

type getOutboundGroup struct {
	key   string
	reply chan int
}

//GetManualNodesMsg is the message type which is used by the rpc server to get the list of persistent peers from the p2p server
type GetManualNodesMsg struct {
	Reply chan []*Peer
}

//DisconnectNodeMsg is the message that is sent to a peer before it gets disconnected
type DisconnectNodeMsg struct {
	Cmp   func(*Peer) bool
	Reply chan error
}

//ConnectNodeMsg is the message type which is used by the rpc server to add a peer to the p2p server
type ConnectNodeMsg struct {
	Addr      string
	Permanent bool
	Reply     chan error
}

//RemoveNodeMsg is the message type which is used by the rpc server to remove a peer from the p2p server
type RemoveNodeMsg struct {
	Cmp   func(*Peer) bool
	Reply chan error
}

// handleQuery is the central handler for all queries and commands from other
// goroutines related to peer state.
func (s *Server) handleQuery(state *peerState, querymsg interface{}) {
	switch msg := querymsg.(type) {
	case getConnCountMsg:
		nconnected := int32(0)
		state.forAllPeers(func(sp *Peer) bool {
			if sp.Connected() {
				nconnected++
			}
			return true
		})
		msg.reply <- nconnected

	case getShouldMineOnGenesisMsg:
		shouldMineOnGenesis := true
		if state.Count() != 0 {
			shouldMineOnGenesis = state.forAllPeers(func(sp *Peer) bool {
				return sp.SelectedTipHash().IsEqual(s.DAGParams.GenesisHash)
			})
		} else {
			shouldMineOnGenesis = false
		}
		msg.reply <- shouldMineOnGenesis

	case GetPeersMsg:
		peers := make([]*Peer, 0, state.Count())
		state.forAllPeers(func(sp *Peer) bool {
			if !sp.Connected() {
				return true
			}
			peers = append(peers, sp)
			return true
		})
		msg.Reply <- peers

	case ConnectNodeMsg:
		// TODO: duplicate oneshots?
		// Limit max number of total peers.
		if state.countOutboundPeers() >= config.ActiveConfig().TargetOutboundPeers {
			msg.Reply <- connmgr.ErrMaxOutboundPeers
			return
		}
		for _, peer := range state.persistentPeers {
			if peer.Addr() == msg.Addr {
				if msg.Permanent {
					msg.Reply <- connmgr.ErrAlreadyConnected
				} else {
					msg.Reply <- connmgr.ErrAlreadyPermanent
				}
				return
			}
		}

		netAddr, err := addrStringToNetAddr(msg.Addr)
		if err != nil {
			msg.Reply <- err
			return
		}

		// TODO: if too many, nuke a non-perm peer.
		spawn(func() {
			s.connManager.Connect(&connmgr.ConnReq{
				Addr:      netAddr,
				Permanent: msg.Permanent,
			})
		})
		msg.Reply <- nil
	case RemoveNodeMsg:
		found := disconnectPeer(state.persistentPeers, msg.Cmp, func(sp *Peer) {
			// Keep group counts ok since we remove from
			// the list now.
			state.outboundGroups[addrmgr.GroupKey(sp.NA())]--
		})

		if found {
			msg.Reply <- nil
		} else {
			msg.Reply <- connmgr.ErrPeerNotFound
		}
	case getOutboundGroup:
		count, ok := state.outboundGroups[msg.key]
		if ok {
			msg.reply <- count
		} else {
			msg.reply <- 0
		}
	// Request a list of the persistent (added) peers.
	case GetManualNodesMsg:
		// Respond with a slice of the relevant peers.
		peers := make([]*Peer, 0, len(state.persistentPeers))
		for _, sp := range state.persistentPeers {
			peers = append(peers, sp)
		}
		msg.Reply <- peers
	case DisconnectNodeMsg:
		// Check inbound peers. We pass a nil callback since we don't
		// require any additional actions on disconnect for inbound peers.
		found := disconnectPeer(state.inboundPeers, msg.Cmp, nil)
		if found {
			msg.Reply <- nil
			return
		}

		// Check outbound peers.
		found = disconnectPeer(state.outboundPeers, msg.Cmp, func(sp *Peer) {
			// Keep group counts ok since we remove from
			// the list now.
			state.outboundGroups[addrmgr.GroupKey(sp.NA())]--
		})
		if found {
			// If there are multiple outbound connections to the same
			// ip:port, continue disconnecting them all until no such
			// peers are found.
			for found {
				found = disconnectPeer(state.outboundPeers, msg.Cmp, func(sp *Peer) {
					state.outboundGroups[addrmgr.GroupKey(sp.NA())]--
				})
			}
			msg.Reply <- nil
			return
		}

		msg.Reply <- connmgr.ErrPeerNotFound
	}
}

// disconnectPeer attempts to drop the connection of a targeted peer in the
// passed peer list. Targets are identified via usage of the passed
// `compareFunc`, which should return `true` if the passed peer is the target
// peer. This function returns true on success and false if the peer is unable
// to be located. If the peer is found, and the passed callback: `whenFound'
// isn't nil, we call it with the peer as the argument before it is removed
// from the peerList, and is disconnected from the server.
func disconnectPeer(peerList map[int32]*Peer, compareFunc func(*Peer) bool, whenFound func(*Peer)) bool {
	for addr, peer := range peerList {
		if compareFunc(peer) {
			if whenFound != nil {
				whenFound(peer)
			}

			// This is ok because we are not continuing
			// to iterate so won't corrupt the loop.
			delete(peerList, addr)
			peer.Disconnect()
			return true
		}
	}
	return false
}

// newPeerConfig returns the configuration for the given serverPeer.
func newPeerConfig(sp *Peer) *peer.Config {
	return &peer.Config{
		Listeners: peer.MessageListeners{
			OnVersion:         sp.OnVersion,
			OnTx:              sp.OnTx,
			OnBlock:           sp.OnBlock,
			OnInv:             sp.OnInv,
			OnGetData:         sp.OnGetData,
			OnGetBlockLocator: sp.OnGetBlockLocator,
			OnBlockLocator:    sp.OnBlockLocator,
			OnGetBlockInvs:    sp.OnGetBlockInvs,
			OnFeeFilter:       sp.OnFeeFilter,
			OnFilterAdd:       sp.OnFilterAdd,
			OnFilterClear:     sp.OnFilterClear,
			OnFilterLoad:      sp.OnFilterLoad,
			OnGetAddr:         sp.OnGetAddr,
			OnAddr:            sp.OnAddr,
			OnGetSelectedTip:  sp.OnGetSelectedTip,
			OnSelectedTip:     sp.OnSelectedTip,
			OnRead:            sp.OnRead,
			OnWrite:           sp.OnWrite,
		},
		SelectedTipHash:   sp.selectedTipHash,
		IsInDAG:           sp.blockExists,
		HostToNetAddress:  sp.server.addrManager.HostToNetAddress,
		Proxy:             config.ActiveConfig().Proxy,
		UserAgentName:     userAgentName,
		UserAgentVersion:  userAgentVersion,
		UserAgentComments: config.ActiveConfig().UserAgentComments,
		DAGParams:         sp.server.DAGParams,
		Services:          sp.server.services,
		DisableRelayTx:    config.ActiveConfig().BlocksOnly,
		ProtocolVersion:   peer.MaxProtocolVersion,
		SubnetworkID:      config.ActiveConfig().SubnetworkID,
	}
}

// inboundPeerConnected is invoked by the connection manager when a new inbound
// connection is established. It initializes a new inbound server peer
// instance, associates it with the connection, and starts a goroutine to wait
// for disconnection.
func (s *Server) inboundPeerConnected(conn net.Conn) {
	sp := newServerPeer(s, false)
	sp.isWhitelisted = isWhitelisted(conn.RemoteAddr())
	sp.Peer = peer.NewInboundPeer(newPeerConfig(sp))
	sp.AssociateConnection(conn)
	spawn(func() {
		s.peerDoneHandler(sp)
	})
}

// outboundPeerConnected is invoked by the connection manager when a new
// outbound connection is established. It initializes a new outbound server
// peer instance, associates it with the relevant state such as the connection
// request instance and the connection itself, and finally notifies the address
// manager of the attempt.
func (s *Server) outboundPeerConnected(state *peerState, msg *outboundPeerConnectedMsg) {
	sp := newServerPeer(s, msg.connReq.Permanent)
	outboundPeer, err := peer.NewOutboundPeer(newPeerConfig(sp), msg.connReq.Addr.String())
	if err != nil {
		srvrLog.Debugf("Cannot create outbound peer %s: %s", msg.connReq.Addr, err)
		s.connManager.Disconnect(msg.connReq.ID())
	}
	sp.Peer = outboundPeer
	sp.connReq = msg.connReq
	sp.isWhitelisted = isWhitelisted(msg.conn.RemoteAddr())
	sp.AssociateConnection(msg.conn)
	spawn(func() {
		s.peerDoneHandler(sp)
	})
	s.addrManager.Attempt(sp.NA())
	state.outboundGroups[addrmgr.GroupKey(sp.NA())]++
}

// peerDoneHandler handles peer disconnects by notifiying the server that it's
// done along with other performing other desirable cleanup.
func (s *Server) peerDoneHandler(sp *Peer) {
	sp.WaitForDisconnect()
	s.donePeers <- sp

	// Only tell sync manager we are gone if we ever told it we existed.
	if sp.VersionKnown() {
		s.SyncManager.DonePeer(sp.Peer)

		// Evict any remaining orphans that were sent by the peer.
		numEvicted := s.TxMemPool.RemoveOrphansByTag(mempool.Tag(sp.ID()))
		if numEvicted > 0 {
			txmpLog.Debugf("Evicted %d %s from peer %s (id %d)",
				numEvicted, logger.PickNoun(numEvicted, "orphan",
					"orphans"), sp, sp.ID())
		}
	}
	close(sp.quit)
}

// peerHandler is used to handle peer operations such as adding and removing
// peers to and from the server, banning peers, and broadcasting messages to
// peers. It must be run in a goroutine.
func (s *Server) peerHandler() {
	// Start the address manager and sync manager, both of which are needed
	// by peers. This is done here since their lifecycle is closely tied
	// to this handler and rather than adding more channels to sychronize
	// things, it's easier and slightly faster to simply start and stop them
	// in this handler.
	s.addrManager.Start()
	s.SyncManager.Start()

	s.quitWaitGroup.Add(1)

	srvrLog.Tracef("Starting peer handler")

	state := &peerState{
		inboundPeers:    make(map[int32]*Peer),
		persistentPeers: make(map[int32]*Peer),
		outboundPeers:   make(map[int32]*Peer),
		banned:          make(map[string]time.Time),
		outboundGroups:  make(map[string]int),
	}

	if !config.ActiveConfig().DisableDNSSeed {
		seedFromSubNetwork := func(subnetworkID *subnetworkid.SubnetworkID) {
			connmgr.SeedFromDNS(config.ActiveConfig().NetParams(), defaultRequiredServices,
				false, subnetworkID, serverutils.KaspadLookup, func(addrs []*wire.NetAddress) {
					// Kaspad uses a lookup of the dns seeder here. Since seeder returns
					// IPs of nodes and not its own IP, we can not know real IP of
					// source. So we'll take first returned address as source.
					s.addrManager.AddAddresses(addrs, addrs[0], subnetworkID)
				})
		}

		// Add full nodes discovered through DNS to the address manager.
		seedFromSubNetwork(nil)

		if config.ActiveConfig().SubnetworkID != nil {
			// Node is partial - fetch nodes with same subnetwork
			seedFromSubNetwork(config.ActiveConfig().SubnetworkID)
		}
	}
	spawn(s.connManager.Start)

out:
	for {
		select {
		// New peers connected to the server.
		case p := <-s.newPeers:
			s.handleAddPeerMsg(state, p)

		// Disconnected peers.
		case p := <-s.donePeers:
			s.handleDonePeerMsg(state, p)

		// Peer to ban.
		case p := <-s.banPeers:
			s.handleBanPeerMsg(state, p)

		// New inventory to potentially be relayed to other peers.
		case invMsg := <-s.relayInv:
			s.handleRelayInvMsg(state, invMsg)

		// Message to broadcast to all connected peers except those
		// which are excluded by the message.
		case bmsg := <-s.broadcast:
			s.handleBroadcastMsg(state, &bmsg)

		case qmsg := <-s.Query:
			s.handleQuery(state, qmsg)

		case <-s.quit:
			// Disconnect all peers on server shutdown.
			state.forAllPeers(func(sp *Peer) bool {
				srvrLog.Tracef("Shutdown peer %s", sp)
				sp.Disconnect()
				return true
			})
			s.quitWaitGroup.Done()
			break out

		case opcMsg := <-s.newOutboundConnection:
			s.outboundPeerConnected(state, opcMsg)
		}
	}

	// Wait for all p2p server quit jobs to finish before stopping the
	// various managers
	s.quitWaitGroup.Wait()

	s.connManager.Stop()
	s.SyncManager.Stop()
	s.addrManager.Stop()

	// Drain channels before exiting so nothing is left waiting around
	// to send.
cleanup:
	for {
		select {
		case <-s.newPeers:
		case <-s.donePeers:
		case <-s.relayInv:
		case <-s.broadcast:
		case <-s.Query:
		default:
			break cleanup
		}
	}
	s.wg.Done()
	srvrLog.Tracef("Peer handler done")
}

// AddPeer adds a new peer that has already been connected to the server.
func (s *Server) AddPeer(sp *Peer) {
	s.newPeers <- sp
}

// BanPeer bans a peer that has already been connected to the server by ip.
func (s *Server) BanPeer(sp *Peer) {
	s.banPeers <- sp
}

// RelayInventory relays the passed inventory vector to all connected peers
// that are not already known to have it.
func (s *Server) RelayInventory(invVect *wire.InvVect, data interface{}) {
	s.relayInv <- relayMsg{invVect: invVect, data: data}
}

// BroadcastMessage sends msg to all peers currently connected to the server
// except those in the passed peers to exclude.
func (s *Server) BroadcastMessage(msg wire.Message, exclPeers ...*Peer) {
	// XXX: Need to determine if this is an alert that has already been
	// broadcast and refrain from broadcasting again.
	bmsg := broadcastMsg{message: msg, excludePeers: exclPeers}
	s.broadcast <- bmsg
}

// ConnectedCount returns the number of currently connected peers.
func (s *Server) ConnectedCount() int32 {
	replyChan := make(chan int32)

	s.Query <- getConnCountMsg{reply: replyChan}

	return <-replyChan
}

// ShouldMineOnGenesis checks if the node is connected to at least one
// peer, and at least one of its peers knows of any blocks that were mined
// on top of the genesis block.
func (s *Server) ShouldMineOnGenesis() bool {
	replyChan := make(chan bool)

	s.Query <- getShouldMineOnGenesisMsg{reply: replyChan}

	return <-replyChan
}

// OutboundGroupCount returns the number of peers connected to the given
// outbound group key.
func (s *Server) OutboundGroupCount(key string) int {
	replyChan := make(chan int)
	s.Query <- getOutboundGroup{key: key, reply: replyChan}
	return <-replyChan
}

// AddBytesSent adds the passed number of bytes to the total bytes sent counter
// for the server. It is safe for concurrent access.
func (s *Server) AddBytesSent(bytesSent uint64) {
	atomic.AddUint64(&s.bytesSent, bytesSent)
}

// AddBytesReceived adds the passed number of bytes to the total bytes received
// counter for the server. It is safe for concurrent access.
func (s *Server) AddBytesReceived(bytesReceived uint64) {
	atomic.AddUint64(&s.bytesReceived, bytesReceived)
}

// NetTotals returns the sum of all bytes received and sent across the network
// for all peers. It is safe for concurrent access.
func (s *Server) NetTotals() (uint64, uint64) {
	return atomic.LoadUint64(&s.bytesReceived),
		atomic.LoadUint64(&s.bytesSent)
}

// rebroadcastHandler keeps track of user submitted inventories that we have
// sent out but have not yet made it into a block. We periodically rebroadcast
// them in case our peers restarted or otherwise lost track of them.
func (s *Server) rebroadcastHandler() {
	// Wait 5 min before first tx rebroadcast.
	timer := time.NewTimer(5 * time.Minute)
	pendingInvs := make(map[wire.InvVect]interface{})

	s.quitWaitGroup.Add(1)

out:
	for {
		select {
		case riv := <-s.modifyRebroadcastInv:
			switch msg := riv.(type) {
			// Incoming InvVects are added to our map of RPC txs.
			case broadcastInventoryAdd:
				pendingInvs[*msg.invVect] = msg.data

			// When an InvVect has been added to a block, we can
			// now remove it, if it was present.
			case broadcastInventoryDel:
				delete(pendingInvs, *msg)
			}

		case <-timer.C:
			// Any inventory we have has not made it into a block
			// yet. We periodically resubmit them until they have.
			for iv, data := range pendingInvs {
				ivCopy := iv
				s.RelayInventory(&ivCopy, data)
			}

			// Process at a random time up to 30mins (in seconds)
			// in the future.
			timer.Reset(time.Second *
				time.Duration(randomUint16Number(1800)))

		case <-s.quit:
			break out
		}
	}

	timer.Stop()

	// Drain channels before exiting so nothing is left waiting around
	// to send.
cleanup:
	for {
		select {
		case <-s.modifyRebroadcastInv:
		default:
			break cleanup
		}
	}
	s.quitWaitGroup.Done()
	s.wg.Done()
}

// Start begins accepting connections from peers.
func (s *Server) Start() {

	// Start the peer handler which in turn starts the address and block
	// managers.
	s.wg.Add(1)
	spawn(s.peerHandler)

	if s.nat != nil {
		s.wg.Add(1)
		spawn(s.upnpUpdateThread)
	}

	cfg := config.ActiveConfig()

	if !cfg.DisableRPC {
		s.wg.Add(1)

		// Start the rebroadcastHandler, which ensures user tx received by
		// the RPC server are rebroadcast until being included in a block.
		spawn(s.rebroadcastHandler)
	}
}

// Stop gracefully shuts down the server by stopping and disconnecting all
// peers and the main listener.
func (s *Server) Stop() error {
	// Signal the remaining goroutines to quit.
	close(s.quit)
	return nil
}

// WaitForShutdown blocks until the main listener and peer handlers are stopped.
func (s *Server) WaitForShutdown() {
	s.wg.Wait()
}

// ScheduleShutdown schedules a server shutdown after the specified duration.
// It also dynamically adjusts how often to warn the server is going down based
// on remaining duration.
func (s *Server) ScheduleShutdown(duration time.Duration) {
	// Don't schedule shutdown more than once.
	if atomic.AddInt32(&s.shutdownSched, 1) != 1 {
		return
	}
	srvrLog.Warnf("Server shutdown in %s", duration)
	spawn(func() {
		remaining := duration
		tickDuration := dynamicTickDuration(remaining)
		done := time.After(remaining)
		ticker := time.NewTicker(tickDuration)
	out:
		for {
			select {
			case <-done:
				ticker.Stop()
				s.Stop()
				break out
			case <-ticker.C:
				remaining = remaining - tickDuration
				if remaining < time.Second {
					continue
				}

				// Change tick duration dynamically based on remaining time.
				newDuration := dynamicTickDuration(remaining)
				if tickDuration != newDuration {
					tickDuration = newDuration
					ticker.Stop()
					ticker = time.NewTicker(tickDuration)
				}
				srvrLog.Warnf("Server shutdown in %s", remaining)
			}
		}
	})
}

// ParseListeners determines whether each listen address is IPv4 and IPv6 and
// returns a slice of appropriate net.Addrs to listen on with TCP. It also
// properly detects addresses which apply to "all interfaces" and adds the
// address as both IPv4 and IPv6.
func ParseListeners(addrs []string) ([]net.Addr, error) {
	netAddrs := make([]net.Addr, 0, len(addrs)*2)
	for _, addr := range addrs {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			// Shouldn't happen due to already being normalized.
			return nil, err
		}

		// Empty host or host of * on plan9 is both IPv4 and IPv6.
		if host == "" || (host == "*" && runtime.GOOS == "plan9") {
			netAddrs = append(netAddrs, simpleAddr{net: "tcp4", addr: addr})
			netAddrs = append(netAddrs, simpleAddr{net: "tcp6", addr: addr})
			continue
		}

		// Strip IPv6 zone id if present since net.ParseIP does not
		// handle it.
		zoneIndex := strings.LastIndex(host, "%")
		if zoneIndex > 0 {
			host = host[:zoneIndex]
		}

		// Parse the IP.
		ip := net.ParseIP(host)
		if ip == nil {
			hostAddrs, err := net.LookupHost(host)
			if err != nil {
				return nil, err
			}
			ip = net.ParseIP(hostAddrs[0])
			if ip == nil {
				return nil, errors.Errorf("Cannot resolve IP address for host '%s'", host)
			}
		}

		// To4 returns nil when the IP is not an IPv4 address, so use
		// this determine the address type.
		if ip.To4() == nil {
			netAddrs = append(netAddrs, simpleAddr{net: "tcp6", addr: addr})
		} else {
			netAddrs = append(netAddrs, simpleAddr{net: "tcp4", addr: addr})
		}
	}
	return netAddrs, nil
}

func (s *Server) upnpUpdateThread() {
	// Go off immediately to prevent code duplication, thereafter we renew
	// lease every 15 minutes.
	timer := time.NewTimer(0 * time.Second)
	lport, _ := strconv.ParseInt(config.ActiveConfig().NetParams().DefaultPort, 10, 16)
	first := true

	s.quitWaitGroup.Add(1)

out:
	for {
		select {
		case <-timer.C:
			// TODO: pick external port  more cleverly
			// TODO: know which ports we are listening to on an external net.
			// TODO: if specific listen port doesn't work then ask for wildcard
			// listen port?
			// XXX this assumes timeout is in seconds.
			listenPort, err := s.nat.AddPortMapping("tcp", int(lport), int(lport),
				"kaspad listen port", 20*60)
			if err != nil {
				srvrLog.Warnf("can't add UPnP port mapping: %s", err)
			}
			if first && err == nil {
				// TODO: look this up periodically to see if upnp domain changed
				// and so did ip.
				externalip, err := s.nat.GetExternalAddress()
				if err != nil {
					srvrLog.Warnf("UPnP can't get external address: %s", err)
					continue out
				}
				na := wire.NewNetAddressIPPort(externalip, uint16(listenPort),
					s.services)
				err = s.addrManager.AddLocalAddress(na, addrmgr.UpnpPrio)
				if err != nil {
					// XXX DeletePortMapping?
				}
				srvrLog.Warnf("Successfully bound via UPnP to %s", addrmgr.NetAddressKey(na))
				first = false
			}
			timer.Reset(time.Minute * 15)
		case <-s.quit:
			break out
		}
	}

	timer.Stop()

	if err := s.nat.DeletePortMapping("tcp", int(lport), int(lport)); err != nil {
		srvrLog.Warnf("unable to remove UPnP port mapping: %s", err)
	} else {
		srvrLog.Debugf("successfully disestablished UPnP port mapping")
	}

	s.quitWaitGroup.Done()
	s.wg.Done()
}

// NewServer returns a new kaspad server configured to listen on addr for the
// kaspa network type specified by dagParams. Use start to begin accepting
// connections from peers.
func NewServer(listenAddrs []string, db database.DB, dagParams *dagconfig.Params, interrupt <-chan struct{}, notifyNewTransactions func(txns []*mempool.TxDesc)) (*Server, error) {
	services := defaultServices
	if config.ActiveConfig().NoPeerBloomFilters {
		services &^= wire.SFNodeBloom
	}

	amgr := addrmgr.New(config.ActiveConfig().DataDir, serverutils.KaspadLookup, config.ActiveConfig().SubnetworkID)

	var listeners []net.Listener
	var nat serverutils.NAT
	if !config.ActiveConfig().DisableListen {
		var err error
		listeners, nat, err = initListeners(amgr, listenAddrs, services)
		if err != nil {
			return nil, err
		}
		if len(listeners) == 0 {
			return nil, errors.New("no valid listen address")
		}
	}

	maxPeers := config.ActiveConfig().TargetOutboundPeers + config.ActiveConfig().MaxInboundPeers

	s := Server{
		DAGParams:             dagParams,
		addrManager:           amgr,
		newPeers:              make(chan *Peer, maxPeers),
		donePeers:             make(chan *Peer, maxPeers),
		banPeers:              make(chan *Peer, maxPeers),
		Query:                 make(chan interface{}),
		relayInv:              make(chan relayMsg, maxPeers),
		broadcast:             make(chan broadcastMsg, maxPeers),
		quit:                  make(chan struct{}),
		modifyRebroadcastInv:  make(chan interface{}),
		newOutboundConnection: make(chan *outboundPeerConnectedMsg, config.ActiveConfig().TargetOutboundPeers),
		nat:                   nat,
		db:                    db,
		TimeSource:            blockdag.NewTimeSource(),
		services:              services,
		SigCache:              txscript.NewSigCache(config.ActiveConfig().SigCacheMaxSize),
		notifyNewTransactions: notifyNewTransactions,
	}

	// Create indexes if needed.
	var indexes []indexers.Indexer
	if config.ActiveConfig().AcceptanceIndex {
		indxLog.Info("acceptance index is enabled")
		s.AcceptanceIndex = indexers.NewAcceptanceIndex()
		indexes = append(indexes, s.AcceptanceIndex)
	}

	// Create an index manager if any of the optional indexes are enabled.
	var indexManager blockdag.IndexManager
	if len(indexes) > 0 {
		indexManager = indexers.NewManager(indexes)
	}

	// Create a new block DAG instance with the appropriate configuration.
	var err error
	s.DAG, err = blockdag.New(&blockdag.Config{
		DB:           s.db,
		Interrupt:    interrupt,
		DAGParams:    s.DAGParams,
		TimeSource:   s.TimeSource,
		SigCache:     s.SigCache,
		IndexManager: indexManager,
		SubnetworkID: config.ActiveConfig().SubnetworkID,
	})
	if err != nil {
		return nil, err
	}

	txC := mempool.Config{
		Policy: mempool.Policy{
			AcceptNonStd:    config.ActiveConfig().RelayNonStd,
			MaxOrphanTxs:    config.ActiveConfig().MaxOrphanTxs,
			MaxOrphanTxSize: config.DefaultMaxOrphanTxSize,
			MinRelayTxFee:   config.ActiveConfig().MinRelayTxFee,
			MaxTxVersion:    1,
		},
		DAGParams:      dagParams,
		MedianTimePast: func() time.Time { return s.DAG.CalcPastMedianTime() },
		CalcSequenceLockNoLock: func(tx *util.Tx, utxoSet blockdag.UTXOSet) (*blockdag.SequenceLock, error) {
			return s.DAG.CalcSequenceLockNoLock(tx, utxoSet, true)
		},
		IsDeploymentActive: s.DAG.IsDeploymentActive,
		SigCache:           s.SigCache,
		DAG:                s.DAG,
	}
	s.TxMemPool = mempool.New(&txC)

	s.SyncManager, err = netsync.New(&netsync.Config{
		PeerNotifier: &s,
		DAG:          s.DAG,
		TxMemPool:    s.TxMemPool,
		DAGParams:    s.DAGParams,
		MaxPeers:     maxPeers,
	})
	if err != nil {
		return nil, err
	}

	// Only setup a function to return new addresses to connect to when
	// not running in connect-only mode. The simulation network is always
	// in connect-only mode since it is only intended to connect to
	// specified peers and actively avoid advertising and connecting to
	// discovered peers in order to prevent it from becoming a public test
	// network.
	var newAddressFunc func() (net.Addr, error)
	if !config.ActiveConfig().Simnet && len(config.ActiveConfig().ConnectPeers) == 0 {
		newAddressFunc = func() (net.Addr, error) {
			for tries := 0; tries < 100; tries++ {
				addr := s.addrManager.GetAddress()
				if addr == nil {
					break
				}

				// Address will not be invalid, local or unroutable
				// because addrmanager rejects those on addition.
				// Just check that we don't already have an address
				// in the same group so that we are not connecting
				// to the same network segment at the expense of
				// others.
				//
				// Networks that accept unroutable connections are exempt
				// from this rule, since they're meant to run within a
				// private subnet, like 10.0.0.0/16.
				if !config.ActiveConfig().NetParams().AcceptUnroutable {
					key := addrmgr.GroupKey(addr.NetAddress())
					if s.OutboundGroupCount(key) != 0 {
						continue
					}
				}

				// only allow recent nodes (10mins) after we failed 30
				// times
				if tries < 30 && time.Since(addr.LastAttempt()) < 10*time.Minute {
					continue
				}

				// allow nondefault ports after 50 failed tries.
				if tries < 50 && fmt.Sprintf("%d", addr.NetAddress().Port) !=
					config.ActiveConfig().NetParams().DefaultPort {
					continue
				}

				addrString := addrmgr.NetAddressKey(addr.NetAddress())
				return addrStringToNetAddr(addrString)
			}
			return nil, connmgr.ErrNoAddress
		}
	}

	// Create a connection manager.
	cmgr, err := connmgr.New(&connmgr.Config{
		Listeners:      listeners,
		OnAccept:       s.inboundPeerConnected,
		RetryDuration:  connectionRetryInterval,
		TargetOutbound: uint32(config.ActiveConfig().TargetOutboundPeers),
		Dial:           serverutils.KaspadDial,
		OnConnection: func(c *connmgr.ConnReq, conn net.Conn) {
			s.newOutboundConnection <- &outboundPeerConnectedMsg{
				connReq: c,
				conn:    conn,
			}
		},
		GetNewAddress: newAddressFunc,
	})
	if err != nil {
		return nil, err
	}
	s.connManager = cmgr

	// Start up persistent peers.
	permanentPeers := config.ActiveConfig().ConnectPeers
	if len(permanentPeers) == 0 {
		permanentPeers = config.ActiveConfig().AddPeers
	}
	for _, addr := range permanentPeers {
		netAddr, err := addrStringToNetAddr(addr)
		if err != nil {
			return nil, err
		}

		spawn(func() {
			s.connManager.Connect(&connmgr.ConnReq{
				Addr:      netAddr,
				Permanent: true,
			})
		})
	}

	return &s, nil
}

// initListeners initializes the configured net listeners and adds any bound
// addresses to the address manager. Returns the listeners and a NAT interface,
// which is non-nil if UPnP is in use.
func initListeners(amgr *addrmgr.AddrManager, listenAddrs []string, services wire.ServiceFlag) ([]net.Listener, serverutils.NAT, error) {
	// Listen for TCP connections at the configured addresses
	netAddrs, err := ParseListeners(listenAddrs)
	if err != nil {
		return nil, nil, err
	}

	listeners := make([]net.Listener, 0, len(netAddrs))
	for _, addr := range netAddrs {
		listener, err := net.Listen(addr.Network(), addr.String())
		if err != nil {
			srvrLog.Warnf("Can't listen on %s: %s", addr, err)
			continue
		}
		listeners = append(listeners, listener)
	}

	var nat serverutils.NAT
	if len(config.ActiveConfig().ExternalIPs) != 0 {
		defaultPort, err := strconv.ParseUint(config.ActiveConfig().NetParams().DefaultPort, 10, 16)
		if err != nil {
			srvrLog.Errorf("Can not parse default port %s for active DAG: %s",
				config.ActiveConfig().NetParams().DefaultPort, err)
			return nil, nil, err
		}

		for _, sip := range config.ActiveConfig().ExternalIPs {
			eport := uint16(defaultPort)
			host, portstr, err := net.SplitHostPort(sip)
			if err != nil {
				// no port, use default.
				host = sip
			} else {
				port, err := strconv.ParseUint(portstr, 10, 16)
				if err != nil {
					srvrLog.Warnf("Can not parse port from %s for "+
						"externalip: %s", sip, err)
					continue
				}
				eport = uint16(port)
			}
			na, err := amgr.HostToNetAddress(host, eport, services)
			if err != nil {
				srvrLog.Warnf("Not adding %s as externalip: %s", sip, err)
				continue
			}

			err = amgr.AddLocalAddress(na, addrmgr.ManualPrio)
			if err != nil {
				amgrLog.Warnf("Skipping specified external IP: %s", err)
			}
		}
	} else {
		if config.ActiveConfig().Upnp {
			var err error
			nat, err = serverutils.Discover()
			if err != nil {
				srvrLog.Warnf("Can't discover upnp: %s", err)
			}
			// nil nat here is fine, just means no upnp on network.
		}

		// Add bound addresses to address manager to be advertised to peers.
		for _, listener := range listeners {
			addr := listener.Addr().String()
			err := addLocalAddress(amgr, addr, services)
			if err != nil {
				amgrLog.Warnf("Skipping bound address %s: %s", addr, err)
			}
		}
	}

	return listeners, nat, nil
}

// addrStringToNetAddr takes an address in the form of 'host:port' and returns
// a net.Addr which maps to the original address with any host names resolved
// to IP addresses. It also handles tor addresses properly by returning a
// net.Addr that encapsulates the address.
func addrStringToNetAddr(addr string) (net.Addr, error) {
	host, strPort, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(strPort)
	if err != nil {
		return nil, err
	}

	// Skip if host is already an IP address.
	if ip := net.ParseIP(host); ip != nil {
		return &net.TCPAddr{
			IP:   ip,
			Port: port,
		}, nil
	}

	// Attempt to look up an IP address associated with the parsed host.
	ips, err := serverutils.KaspadLookup(host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, errors.Errorf("no addresses found for %s", host)
	}

	return &net.TCPAddr{
		IP:   ips[0],
		Port: port,
	}, nil
}

// addLocalAddress adds an address that this node is listening on to the
// address manager so that it may be relayed to peers.
func addLocalAddress(addrMgr *addrmgr.AddrManager, addr string, services wire.ServiceFlag) error {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return err
	}

	if ip := net.ParseIP(host); ip != nil && ip.IsUnspecified() {
		// If bound to unspecified address, advertise all local interfaces
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			return err
		}

		for _, addr := range addrs {
			ifaceIP, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				continue
			}

			// If bound to 0.0.0.0, do not add IPv6 interfaces and if bound to
			// ::, do not add IPv4 interfaces.
			if (ip.To4() == nil) != (ifaceIP.To4() == nil) {
				continue
			}

			netAddr := wire.NewNetAddressIPPort(ifaceIP, uint16(port), services)
			addrMgr.AddLocalAddress(netAddr, addrmgr.BoundPrio)
		}
	} else {
		netAddr, err := addrMgr.HostToNetAddress(host, uint16(port), services)
		if err != nil {
			return err
		}

		addrMgr.AddLocalAddress(netAddr, addrmgr.BoundPrio)
	}

	return nil
}

// dynamicTickDuration is a convenience function used to dynamically choose a
// tick duration based on remaining time. It is primarily used during
// server shutdown to make shutdown warnings more frequent as the shutdown time
// approaches.
func dynamicTickDuration(remaining time.Duration) time.Duration {
	switch {
	case remaining <= time.Second*5:
		return time.Second
	case remaining <= time.Second*15:
		return time.Second * 5
	case remaining <= time.Minute:
		return time.Second * 15
	case remaining <= time.Minute*5:
		return time.Minute
	case remaining <= time.Minute*15:
		return time.Minute * 5
	case remaining <= time.Hour:
		return time.Minute * 15
	}
	return time.Hour
}

// isWhitelisted returns whether the IP address is included in the whitelisted
// networks and IPs.
func isWhitelisted(addr net.Addr) bool {
	if len(config.ActiveConfig().Whitelists) == 0 {
		return false
	}

	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		srvrLog.Warnf("Unable to SplitHostPort on '%s': %s", addr, err)
		return false
	}
	ip := net.ParseIP(host)
	if ip == nil {
		srvrLog.Warnf("Unable to parse IP '%s'", addr)
		return false
	}

	for _, ipnet := range config.ActiveConfig().Whitelists {
		if ipnet.Contains(ip) {
			return true
		}
	}
	return false
}

// AnnounceNewTransactions generates and relays inventory vectors and notifies
// both websocket and getblocktemplate long poll clients of the passed
// transactions. This function should be called whenever new transactions
// are added to the mempool.
func (s *Server) AnnounceNewTransactions(txns []*mempool.TxDesc) {
	// Generate and relay inventory vectors for all newly accepted
	// transactions.
	s.RelayTransactions(txns)

	// Notify both websocket and getblocktemplate long poll clients of all
	// newly accepted transactions.
	s.notifyNewTransactions(txns)
}

// TransactionConfirmed is a function for the peerNotifier interface.
// When a transaction has one confirmation, we can mark it as no
// longer needing rebroadcasting.
func (s *Server) TransactionConfirmed(tx *util.Tx) {
	// Rebroadcasting is only necessary when the RPC server is active.
	if config.ActiveConfig().DisableRPC {
		return
	}

	iv := wire.NewInvVect(wire.InvTypeTx, (*daghash.Hash)(tx.ID()))
	s.RemoveRebroadcastInventory(iv)
}
