package netadapter

import (
	"sync"
	"sync/atomic"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/id"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server/grpcserver"
	"github.com/pkg/errors"
)

// RouterInitializer is a function that initializes a new
// router to be used with a new connection
type RouterInitializer func(*routerpkg.Router, *NetConnection)

// NetAdapter is an abstraction layer over networking.
// This type expects a RouteInitializer function. This
// function weaves together the various "routes" (messages
// and message handlers) without exposing anything related
// to networking internals.
type NetAdapter struct {
	cfg                  *config.Config
	id                   *id.ID
	p2pServer            server.P2PServer
	p2pRouterInitializer RouterInitializer
	rpcServer            server.Server
	rpcRouterInitializer RouterInitializer
	stop                 uint32

	p2pConnections     map[*NetConnection]struct{}
	p2pConnectionsLock sync.RWMutex
}

// NewNetAdapter creates and starts a new NetAdapter on the
// given listeningPort
func NewNetAdapter(cfg *config.Config) (*NetAdapter, error) {
	netAdapterID, err := id.GenerateID()
	if err != nil {
		return nil, err
	}
	p2pServer, err := grpcserver.NewP2PServer(cfg.Listeners)
	if err != nil {
		return nil, err
	}
	rpcServer, err := grpcserver.NewRPCServer(cfg.RPCListeners)
	if err != nil {
		return nil, err
	}
	adapter := NetAdapter{
		cfg:       cfg,
		id:        netAdapterID,
		p2pServer: p2pServer,
		rpcServer: rpcServer,

		p2pConnections: make(map[*NetConnection]struct{}),
	}

	adapter.p2pServer.SetOnConnectedHandler(adapter.onP2PConnectedHandler)
	adapter.rpcServer.SetOnConnectedHandler(adapter.onRPCConnectedHandler)

	return &adapter, nil
}

// Start begins the operation of the NetAdapter
func (na *NetAdapter) Start() error {
	if na.p2pRouterInitializer == nil {
		return errors.New("p2pRouterInitializer was not set")
	}
	if na.rpcRouterInitializer == nil {
		return errors.New("rpcRouterInitializer was not set")
	}

	err := na.p2pServer.Start()
	if err != nil {
		return err
	}
	err = na.rpcServer.Start()
	if err != nil {
		return err
	}

	return nil
}

// Stop safely closes the NetAdapter
func (na *NetAdapter) Stop() error {
	if atomic.AddUint32(&na.stop, 1) != 1 {
		return errors.New("net adapter stopped more than once")
	}
	err := na.p2pServer.Stop()
	if err != nil {
		return err
	}
	return na.rpcServer.Stop()
}

// P2PConnect tells the NetAdapter's underlying p2p server to initiate a connection
// to the given address
func (na *NetAdapter) P2PConnect(address string) error {
	_, err := na.p2pServer.Connect(address)
	return err
}

// P2PConnections returns a list of p2p connections currently connected and active
func (na *NetAdapter) P2PConnections() []*NetConnection {
	na.p2pConnectionsLock.RLock()
	defer na.p2pConnectionsLock.RUnlock()

	netConnections := make([]*NetConnection, 0, len(na.p2pConnections))

	for netConnection := range na.p2pConnections {
		netConnections = append(netConnections, netConnection)
	}

	return netConnections
}

// P2PConnectionCount returns the count of the connected p2p connections
func (na *NetAdapter) P2PConnectionCount() int {
	na.p2pConnectionsLock.RLock()
	defer na.p2pConnectionsLock.RUnlock()

	return len(na.p2pConnections)
}

func (na *NetAdapter) onP2PConnectedHandler(connection server.Connection) error {
	netConnection := newNetConnection(connection, na.p2pRouterInitializer)

	na.p2pConnectionsLock.Lock()
	defer na.p2pConnectionsLock.Unlock()

	netConnection.setOnDisconnectedHandler(func() {
		na.p2pConnectionsLock.Lock()
		defer na.p2pConnectionsLock.Unlock()

		delete(na.p2pConnections, netConnection)
	})

	na.p2pConnections[netConnection] = struct{}{}

	netConnection.start()

	return nil
}

func (na *NetAdapter) onRPCConnectedHandler(connection server.Connection) error {
	netConnection := newNetConnection(connection, na.rpcRouterInitializer)
	netConnection.setOnDisconnectedHandler(func() {})
	netConnection.start()

	return nil
}

// SetP2PRouterInitializer sets the p2pRouterInitializer function
// for the net adapter
func (na *NetAdapter) SetP2PRouterInitializer(routerInitializer RouterInitializer) {
	na.p2pRouterInitializer = routerInitializer
}

// SetRPCRouterInitializer sets the rpcRouterInitializer function
// for the net adapter
func (na *NetAdapter) SetRPCRouterInitializer(routerInitializer RouterInitializer) {
	na.rpcRouterInitializer = routerInitializer
}

// ID returns this netAdapter's ID in the network
func (na *NetAdapter) ID() *id.ID {
	return na.id
}

// P2PBroadcast sends the given `message` to every peer corresponding
// to each NetConnection in the given netConnections
func (na *NetAdapter) P2PBroadcast(netConnections []*NetConnection, message appmessage.Message) error {
	na.p2pConnectionsLock.RLock()
	defer na.p2pConnectionsLock.RUnlock()

	for _, netConnection := range netConnections {
		err := netConnection.router.OutgoingRoute().Enqueue(message)
		if err != nil {
			if errors.Is(err, routerpkg.ErrRouteClosed) {
				log.Debugf("Cannot enqueue message to %s: router is closed", netConnection)
				continue
			}
			return err
		}
	}
	return nil
}
