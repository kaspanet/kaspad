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
}

// New instantiates a new instance of a ConnectionManager
func New(netAdapter *netadapter.NetAdapter, addressManager *addrmgr.AddrManager) (*ConnectionManager, error) {
	c := &ConnectionManager{
		netAdapter:       netAdapter,
		addressManager:   addressManager,
		activeRequested:  map[string]*connectionRequest{},
		pendingRequested: map[string]*connectionRequest{},
		activeOutgoing:   map[string]struct{}{},
		activeIncoming:   map[string]struct{}{},
	}

	cfg := config.ActiveConfig()
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
	spawn(c.connectionsLoop)
}

// Stop halts the operation of the ConnectionManager
func (c *ConnectionManager) Stop() {
	atomic.StoreUint32(&c.stop, 1)

	for _, connection := range c.netAdapter.Connections() {
		_ = c.netAdapter.Disconnect(connection) // Ignore errors since connection might be in the midst of disconnecting
	}
}

const connectionsLoopInterval = 30 * time.Second

func (c *ConnectionManager) initiateConnection(address string) error {
	log.Infof("Connecting to %s", address)
	_, err := c.netAdapter.Connect(address)
	return err
}

func (c *ConnectionManager) connectionsLoop() {
	for atomic.LoadUint32(&c.stop) == 0 {
		connections := c.netAdapter.Connections()

		// We convert the connections list to a set, so that connections can be found quickly
		// Then we go over the set, classifying connection by category: requested, outgoing or incoming
		// Every step removes all matching connections so that once we get to checkIncomingConnections -
		// the only connections left are the incoming ones
		connSet := convertToSet(connections)

		c.checkRequestedConnections(connSet)

		c.checkOutgoingConnections(connSet)

		c.checkIncomingConnections(connSet)

		<-time.Tick(connectionsLoopInterval)
	}
}
