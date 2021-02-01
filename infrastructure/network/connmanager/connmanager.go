package connmanager

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/kaspanet/kaspad/infrastructure/network/addressmanager"

	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"

	"github.com/kaspanet/kaspad/infrastructure/config"
)

// connectionRequest represents a user request (either through CLI or RPC) to connect to a certain node
type connectionRequest struct {
	address       string
	isPermanent   bool
	nextAttempt   time.Time
	retryDuration time.Duration
}

// ConnectionManager monitors that the current active connections satisfy the requirements of
// outgoing, requested and incoming connections
type ConnectionManager struct {
	cfg            *config.Config
	netAdapter     *netadapter.NetAdapter
	addressManager *addressmanager.AddressManager

	activeRequested  map[string]*connectionRequest
	pendingRequested map[string]*connectionRequest
	activeOutgoing   map[string]struct{}
	targetOutgoing   int
	activeIncoming   map[string]struct{}
	maxIncoming      int

	stop                   uint32
	connectionRequestsLock sync.Mutex

	resetLoopChan chan struct{}
	loopTicker    *time.Ticker
}

// New instantiates a new instance of a ConnectionManager
func New(cfg *config.Config, netAdapter *netadapter.NetAdapter, addressManager *addressmanager.AddressManager) (*ConnectionManager, error) {
	c := &ConnectionManager{
		cfg:              cfg,
		netAdapter:       netAdapter,
		addressManager:   addressManager,
		activeRequested:  map[string]*connectionRequest{},
		pendingRequested: map[string]*connectionRequest{},
		activeOutgoing:   map[string]struct{}{},
		activeIncoming:   map[string]struct{}{},
		resetLoopChan:    make(chan struct{}),
		loopTicker:       time.NewTicker(connectionsLoopInterval),
	}

	connectPeers := cfg.AddPeers
	if len(cfg.ConnectPeers) > 0 {
		connectPeers = cfg.ConnectPeers
	}

	c.maxIncoming = cfg.MaxInboundPeers
	c.targetOutgoing = cfg.TargetOutboundPeers

	for _, connectPeer := range connectPeers {
		c.pendingRequested[connectPeer] = &connectionRequest{
			address:     connectPeer,
			isPermanent: true,
		}
	}

	return c, nil
}

// Start begins the operation of the ConnectionManager
func (c *ConnectionManager) Start() {
	spawn("ConnectionManager.connectionsLoop", c.connectionsLoop)
}

// Stop halts the operation of the ConnectionManager
func (c *ConnectionManager) Stop() {
	atomic.StoreUint32(&c.stop, 1)

	for _, connection := range c.netAdapter.P2PConnections() {
		connection.Disconnect()
	}

	c.loopTicker.Stop()
}

func (c *ConnectionManager) run() {
	c.resetLoopChan <- struct{}{}
}

func (c *ConnectionManager) initiateConnection(address string) error {
	log.Infof("Connecting to %s", address)
	return c.netAdapter.P2PConnect(address)
}

const connectionsLoopInterval = 30 * time.Second

func (c *ConnectionManager) connectionsLoop() {
	for atomic.LoadUint32(&c.stop) == 0 {
		connections := c.netAdapter.P2PConnections()

		// We convert the connections list to a set, so that connections can be found quickly
		// Then we go over the set, classifying connection by category: requested, outgoing or incoming.
		// Every step removes all matching connections so that once we get to checkIncomingConnections -
		// the only connections left are the incoming ones
		connSet := convertToSet(connections)

		c.checkRequestedConnections(connSet)

		c.checkOutgoingConnections(connSet)

		c.checkIncomingConnections(connSet)

		c.waitTillNextIteration()
	}
}

// ConnectionCount returns the count of the connected connections
func (c *ConnectionManager) ConnectionCount() int {
	return c.netAdapter.P2PConnectionCount()
}

// Ban marks the given netConnection as banned
func (c *ConnectionManager) Ban(netConnection *netadapter.NetConnection) {
	if c.isPermanent(netConnection.Address()) {
		log.Infof("Cannot ban %s because it's a permanent connection", netConnection.Address())
		return
	}

	c.addressManager.Ban(netConnection.NetAddress())
}

// IsBanned returns whether the given netConnection is banned
func (c *ConnectionManager) IsBanned(netConnection *netadapter.NetConnection) (bool, error) {
	if c.isPermanent(netConnection.Address()) {
		return false, nil
	}

	return c.addressManager.IsBanned(netConnection.NetAddress())
}

func (c *ConnectionManager) waitTillNextIteration() {
	select {
	case <-c.resetLoopChan:
		c.loopTicker.Reset(connectionsLoopInterval)
	case <-c.loopTicker.C:
	}
}

func (c *ConnectionManager) connectionExists(addressString string) bool {
	if _, ok := c.activeRequested[addressString]; ok {
		return true
	}

	if _, ok := c.activeOutgoing[addressString]; ok {
		return true
	}

	if _, ok := c.activeIncoming[addressString]; ok {
		return true
	}

	return false
}

func (c *ConnectionManager) isPermanent(addressString string) bool {
	if conn, ok := c.activeRequested[addressString]; ok {
		return conn.isPermanent
	}

	if conn, ok := c.pendingRequested[addressString]; ok {
		return conn.isPermanent
	}

	return false
}
