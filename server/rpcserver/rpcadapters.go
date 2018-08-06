// Copyright (c) 2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package rpcserver

import (
	"sync/atomic"

	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/mempool"
	"github.com/daglabs/btcd/netsync"
	"github.com/daglabs/btcd/peer"
	"github.com/daglabs/btcd/server"
	"github.com/daglabs/btcd/wire"
	"github.com/daglabs/btcutil"
)

// rpcPeer provides a peer for use with the RPC server and implements the
// rpcserverPeer interface.
type rpcPeer server.Peer

// Ensure rpcPeer implements the rpcserverPeer interface.
var _ server.RPCServerPeer = (*rpcPeer)(nil)

// ToPeer returns the underlying peer instance.
//
// This function is safe for concurrent access and is part of the rpcserverPeer
// interface implementation.
func (p *rpcPeer) ToPeer() *peer.Peer {
	if p == nil {
		return nil
	}
	return (*server.Peer)(p).Peer
}

// IsTxRelayDisabled returns whether or not the peer has disabled transaction
// relay.
//
// This function is safe for concurrent access and is part of the rpcserverPeer
// interface implementation.
func (p *rpcPeer) IsTxRelayDisabled() bool {
	return (*server.Peer)(p).DisableRelayTx
}

// BanScore returns the current integer value that represents how close the peer
// is to being banned.
//
// This function is safe for concurrent access and is part of the rpcserverPeer
// interface implementation.
func (p *rpcPeer) BanScore() uint32 {
	return (*server.Peer)(p).DynamicBanScore.Int()
}

// FeeFilter returns the requested current minimum fee rate for which
// transactions should be announced.
//
// This function is safe for concurrent access
func FeeFilter(p *rpcPeer) int64 {
	return atomic.LoadInt64(&(*server.Peer)(p).NonSafeFeeFilter)
}

// rpcConnManager provides a connection manager for use with the RPC server and
// implements the rpcserverConnManager interface.
type rpcConnManager struct {
	server *server.Server
}

// Ensure rpcConnManager implements the rpcserverConnManager interface.
var _ server.RPCServerConnManager = &rpcConnManager{}

// Connect adds the provided address as a new outbound peer.  The permanent flag
// indicates whether or not to make the peer persistent and reconnect if the
// connection is lost.  Attempting to connect to an already existing peer will
// return an error.
//
// This function is safe for concurrent access and is part of the
// rpcserverConnManager interface implementation.
func (cm *rpcConnManager) Connect(addr string, permanent bool) error {
	replyChan := make(chan error)
	cm.server.Query <- server.ConnectNodeMsg{
		Addr:      addr,
		Permanent: permanent,
		Reply:     replyChan,
	}
	return <-replyChan
}

// RemoveByID removes the peer associated with the provided id from the list of
// persistent peers.  Attempting to remove an id that does not exist will return
// an error.
//
// This function is safe for concurrent access and is part of the
// rpcserverConnManager interface implementation.
func (cm *rpcConnManager) RemoveByID(id int32) error {
	replyChan := make(chan error)
	cm.server.Query <- server.RemoveNodeMsg{
		Cmp:   func(sp *server.Peer) bool { return sp.ID() == id },
		Reply: replyChan,
	}
	return <-replyChan
}

// RemoveByAddr removes the peer associated with the provided address from the
// list of persistent peers.  Attempting to remove an address that does not
// exist will return an error.
//
// This function is safe for concurrent access and is part of the
// rpcserverConnManager interface implementation.
func (cm *rpcConnManager) RemoveByAddr(addr string) error {
	replyChan := make(chan error)
	cm.server.Query <- server.RemoveNodeMsg{
		Cmp:   func(sp *server.Peer) bool { return sp.Addr() == addr },
		Reply: replyChan,
	}
	return <-replyChan
}

// DisconnectByID disconnects the peer associated with the provided id.  This
// applies to both inbound and outbound peers.  Attempting to remove an id that
// does not exist will return an error.
//
// This function is safe for concurrent access and is part of the
// rpcserverConnManager interface implementation.
func (cm *rpcConnManager) DisconnectByID(id int32) error {
	replyChan := make(chan error)
	cm.server.Query <- server.DisconnectNodeMsg{
		Cmp:   func(sp *server.Peer) bool { return sp.ID() == id },
		Reply: replyChan,
	}
	return <-replyChan
}

// DisconnectByAddr disconnects the peer associated with the provided address.
// This applies to both inbound and outbound peers.  Attempting to remove an
// address that does not exist will return an error.
//
// This function is safe for concurrent access and is part of the
// rpcserverConnManager interface implementation.
func (cm *rpcConnManager) DisconnectByAddr(addr string) error {
	replyChan := make(chan error)
	cm.server.Query <- server.DisconnectNodeMsg{
		Cmp:   func(sp *server.Peer) bool { return sp.Addr() == addr },
		Reply: replyChan,
	}
	return <-replyChan
}

// ConnectedCount returns the number of currently connected peers.
//
// This function is safe for concurrent access and is part of the
// rpcserverConnManager interface implementation.
func (cm *rpcConnManager) ConnectedCount() int32 {
	return cm.server.ConnectedCount()
}

// NetTotals returns the sum of all bytes received and sent across the network
// for all peers.
//
// This function is safe for concurrent access and is part of the
// rpcserverConnManager interface implementation.
func (cm *rpcConnManager) NetTotals() (uint64, uint64) {
	return cm.server.NetTotals()
}

// ConnectedPeers returns an array consisting of all connected peers.
//
// This function is safe for concurrent access and is part of the
// rpcserverConnManager interface implementation.
func (cm *rpcConnManager) ConnectedPeers() []server.RPCServerPeer {
	replyChan := make(chan []*server.Peer)
	cm.server.Query <- server.GetPeersMsg{Reply: replyChan}
	serverPeers := <-replyChan

	// Convert to RPC server peers.
	peers := make([]server.RPCServerPeer, 0, len(serverPeers))
	for _, sp := range serverPeers {
		peers = append(peers, (*rpcPeer)(sp))
	}
	return peers
}

// PersistentPeers returns an array consisting of all the added persistent
// peers.
//
// This function is safe for concurrent access and is part of the
// rpcserverConnManager interface implementation.
func (cm *rpcConnManager) PersistentPeers() []server.RPCServerPeer {
	replyChan := make(chan []*server.Peer)
	cm.server.Query <- server.GetAddedNodesMsg{Reply: replyChan}
	serverPeers := <-replyChan

	// Convert to generic peers.
	peers := make([]server.RPCServerPeer, 0, len(serverPeers))
	for _, sp := range serverPeers {
		peers = append(peers, (*rpcPeer)(sp))
	}
	return peers
}

// BroadcastMessage sends the provided message to all currently connected peers.
//
// This function is safe for concurrent access and is part of the
// rpcserverConnManager interface implementation.
func (cm *rpcConnManager) BroadcastMessage(msg wire.Message) {
	cm.server.BroadcastMessage(msg)
}

// AddRebroadcastInventory adds the provided inventory to the list of
// inventories to be rebroadcast at random intervals until they show up in a
// block.
//
// This function is safe for concurrent access and is part of the
// rpcserverConnManager interface implementation.
func (cm *rpcConnManager) AddRebroadcastInventory(iv *wire.InvVect, data interface{}) {
	cm.server.AddRebroadcastInventory(iv, data)
}

// RelayTransactions generates and relays inventory vectors for all of the
// passed transactions to all connected peers.
func (cm *rpcConnManager) RelayTransactions(txns []*mempool.TxDesc) {
	cm.server.RelayTransactions(txns)
}

// rpcSyncMgr provides a block manager for use with the RPC server and
// implements the rpcserverSyncManager interface.
type rpcSyncMgr struct {
	server  *server.Server
	syncMgr *netsync.SyncManager
}

// Ensure rpcSyncMgr implements the rpcserverSyncManager interface.
var _ server.RPCServerSyncManager = (*rpcSyncMgr)(nil)

// IsCurrent returns whether or not the sync manager believes the chain is
// current as compared to the rest of the network.
//
// This function is safe for concurrent access and is part of the
// server.RPCServerSyncManager interface implementation.
func (b *rpcSyncMgr) IsCurrent() bool {
	return b.syncMgr.IsCurrent()
}

// SubmitBlock submits the provided block to the network after processing it
// locally.
//
// This function is safe for concurrent access and is part of the
// server.RPCServerSyncManager interface implementation.
func (b *rpcSyncMgr) SubmitBlock(block *btcutil.Block, flags blockdag.BehaviorFlags) (bool, error) {
	return b.syncMgr.ProcessBlock(block, flags)
}

// Pause pauses the sync manager until the returned channel is closed.
//
// This function is safe for concurrent access and is part of the
// server.RPCServerSyncManager interface implementation.
func (b *rpcSyncMgr) Pause() chan<- struct{} {
	return b.syncMgr.Pause()
}

// SyncPeerID returns the peer that is currently the peer being used to sync
// from.
//
// This function is safe for concurrent access and is part of the
// server.RPCServerSyncManager interface implementation.
func (b *rpcSyncMgr) SyncPeerID() int32 {
	return b.syncMgr.SyncPeerID()
}

// LocateBlocks returns the hashes of the blocks after the first known block in
// the provided locators until the provided stop hash or the current tip is
// reached, up to a max of wire.MaxBlockHeadersPerMsg hashes.
//
// This function is safe for concurrent access and is part of the
// server.rpcserverSyncManager interface implementation.
func (b *rpcSyncMgr) LocateHeaders(locators []*daghash.Hash, hashStop *daghash.Hash) []wire.BlockHeader {
	return b.server.DAG.LocateHeaders(locators, hashStop)
}
