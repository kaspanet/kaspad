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

type ConnectionManager struct {
	netAdapter                *netadapter.NetAdapter
	activeConnectionRequests  map[string]*connectionRequest
	pendingConnectionRequests map[string]*connectionRequest
	activeOutgoing            map[string]struct{}
	targetOutgoing            int
	activeIncoming            map[string]struct{}
	maxIncoming               int
	addressManager            *addrmgr.AddrManager

	stop                   uint32
	connectionRequestsLock sync.Mutex
}

func New(netAdapter *netadapter.NetAdapter, addressManager *addrmgr.AddrManager) (*ConnectionManager, error) {
	c := &ConnectionManager{
		netAdapter:                netAdapter,
		addressManager:            addressManager,
		activeConnectionRequests:  map[string]*connectionRequest{},
		pendingConnectionRequests: map[string]*connectionRequest{},
		activeOutgoing:            map[string]struct{}{},
		activeIncoming:            map[string]struct{}{},
	}

	cfg := config.ActiveConfig()
	connectPeers := cfg.AddPeers
	if len(cfg.ConnectPeers) > 0 {
		connectPeers = cfg.ConnectPeers
	}

	c.maxIncoming = cfg.MaxInboundPeers
	c.targetOutgoing = cfg.TargetOutboundPeers

	for _, connectPeer := range connectPeers {
		c.pendingConnectionRequests[connectPeer] = &connectionRequest{
			address:     connectPeer,
			isPermanent: true,
		}
	}

	return c, nil
}

func (c *ConnectionManager) Start() {
	spawn(c.connectionsLoop)
}

func (c *ConnectionManager) Stop() {
	atomic.StoreUint32(&c.stop, 1)
}

const connectionsLoopInterval = 30 * time.Second

func (c *ConnectionManager) initiateConnection(address string) error {
	log.Infof("Connecting to %s", address)
	_, err := c.netAdapter.Connect(address)
	if err != nil {
		log.Infof("Couldn't connect to %s: %s", address, err)
	}
	return err
}

func (c *ConnectionManager) connectionsLoop() {
	for atomic.LoadUint32(&c.stop) == 0 {
		connections := c.netAdapter.Connections()

		connSet := convertToSet(connections)

		c.checkConnectionRequests(connSet)

		c.checkOutgoingConnections(connSet)

		c.checkIncomingConnections(connSet)

		<-time.Tick(connectionsLoopInterval)
	}
}

// checkIncomingConnections makes sure there's no more then maxIncoming incoming connections
// if there are - it randomly disconnects enough to go below that number
func (c *ConnectionManager) checkIncomingConnections(connSet connectionSet) {
	if len(connSet) <= c.maxIncoming {
		return
	}

	// randomly disconnect nodes until the number of incoming connections is smaller the maxIncoming
	for address, connection := range connSet {
		err := connection.Disconnect()
		if err != nil {
			log.Errorf("Error disconnecting from %s: %+v", address, err)
		}

		connSet.remove(connection)
	}
}
