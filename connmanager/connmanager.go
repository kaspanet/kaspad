package connmanager

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/kaspanet/kaspad/addrmgr"
	"github.com/kaspanet/kaspad/netadapter"

	"github.com/kaspanet/kaspad/config"
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
	addressManager *addrmgr.AddrManager

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
func New(cfg *config.Config, netAdapter *netadapter.NetAdapter, addressManager *addrmgr.AddrManager) (*ConnectionManager, error) {
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

	for _, connection := range c.netAdapter.Connections() {
		_ = c.netAdapter.Disconnect(connection) // Ignore errors since connection might be in the midst of disconnecting
	}

	c.loopTicker.Stop()
}

func (c *ConnectionManager) run() {
	c.resetLoopChan <- struct{}{}
}

func (c *ConnectionManager) initiateConnection(address string) error {
	log.Infof("Connecting to %s", address)
	return c.netAdapter.Connect(address)
}

const connectionsLoopInterval = 30 * time.Second

func (c *ConnectionManager) connectionsLoop() {
	for atomic.LoadUint32(&c.stop) == 0 {
		connections := c.netAdapter.Connections()

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
	return c.netAdapter.ConnectionCount()
}

// Ban prevents the given netConnection from connecting again
func (c *ConnectionManager) Ban(netConnection *netadapter.NetConnection) {
	c.netAdapter.Ban(netConnection)
}

func (c *ConnectionManager) waitTillNextIteration() {
	select {
	case <-c.resetLoopChan:
		c.loopTicker.Stop()
		c.loopTicker = time.NewTicker(connectionsLoopInterval)
	case <-c.loopTicker.C:
	}
}
