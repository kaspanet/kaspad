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
	stop                 uint32

	connections     map[*NetConnection]struct{}
	connectionsLock sync.RWMutex
}

// NewNetAdapter creates and starts a new NetAdapter on the
// given listeningPort
func NewNetAdapter(cfg *config.Config) (*NetAdapter, error) {
	netAdapterID, err := id.GenerateID()
	if err != nil {
		return nil, err
	}
	s, err := grpcserver.NewP2PServer(cfg.Listeners)
	if err != nil {
		return nil, err
	}
	adapter := NetAdapter{
		cfg:       cfg,
		id:        netAdapterID,
		p2pServer: s,

		connections: make(map[*NetConnection]struct{}),
	}

	adapter.p2pServer.SetOnConnectedHandler(adapter.onConnectedHandler)

	return &adapter, nil
}

// Start begins the operation of the NetAdapter
func (na *NetAdapter) Start() error {
	if na.p2pRouterInitializer == nil {
		return errors.New("p2pRouterInitializer was not set")
	}

	err := na.p2pServer.Start()
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
	return na.p2pServer.Stop()
}

// Connect tells the NetAdapter's underlying p2p server to initiate a connection
// to the given address
func (na *NetAdapter) Connect(address string) error {
	_, err := na.p2pServer.Connect(address)
	return err
}

// Connections returns a list of connections currently connected and active
func (na *NetAdapter) Connections() []*NetConnection {
	na.connectionsLock.RLock()
	defer na.connectionsLock.RUnlock()

	netConnections := make([]*NetConnection, 0, len(na.connections))

	for netConnection := range na.connections {
		netConnections = append(netConnections, netConnection)
	}

	return netConnections
}

// ConnectionCount returns the count of the connected connections
func (na *NetAdapter) ConnectionCount() int {
	na.connectionsLock.RLock()
	defer na.connectionsLock.RUnlock()

	return len(na.connections)
}

func (na *NetAdapter) onConnectedHandler(connection server.Connection) error {
	netConnection := newNetConnection(connection, na.p2pRouterInitializer)

	na.connectionsLock.Lock()
	defer na.connectionsLock.Unlock()

	netConnection.setOnDisconnectedHandler(func() {
		na.connectionsLock.Lock()
		defer na.connectionsLock.Unlock()

		delete(na.connections, netConnection)
	})

	na.connections[netConnection] = struct{}{}

	netConnection.start()

	return nil
}

// SetP2PRouterInitializer sets the p2pRouterInitializer function
// for the net adapter
func (na *NetAdapter) SetP2PRouterInitializer(routerInitializer RouterInitializer) {
	na.p2pRouterInitializer = routerInitializer
}

// ID returns this netAdapter's ID in the network
func (na *NetAdapter) ID() *id.ID {
	return na.id
}

// Broadcast sends the given `message` to every peer corresponding
// to each NetConnection in the given netConnections
func (na *NetAdapter) Broadcast(netConnections []*NetConnection, message appmessage.Message) error {
	na.connectionsLock.RLock()
	defer na.connectionsLock.RUnlock()

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
