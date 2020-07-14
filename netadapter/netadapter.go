package netadapter

import (
	"net"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/netadapter/id"
	routerpkg "github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/netadapter/server"
	"github.com/kaspanet/kaspad/netadapter/server/grpcserver"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

// RouterInitializer is a function that initializes a new
// router to be used with a new connection
type RouterInitializer func() (*routerpkg.Router, error)

// NetAdapter is an abstraction layer over networking.
// This type expects a RouteInitializer function. This
// function weaves together the various "routes" (messages
// and message handlers) without exposing anything related
// to networking internals.
type NetAdapter struct {
	id                *id.ID
	server            server.Server
	routerInitializer RouterInitializer
	stop              uint32

	routersToConnections map[*routerpkg.Router]server.Connection
	connectionsToIDs     map[server.Connection]*id.ID
	idsToRouters         map[*id.ID]*routerpkg.Router
	sync.RWMutex
}

// NewNetAdapter creates and starts a new NetAdapter on the
// given listeningPort
func NewNetAdapter(listeningAddrs []string) (*NetAdapter, error) {
	netAdapterID, err := id.GenerateID()
	if err != nil {
		return nil, err
	}
	s, err := grpcserver.NewGRPCServer(listeningAddrs)
	if err != nil {
		return nil, err
	}
	adapter := NetAdapter{
		id:     netAdapterID,
		server: s,

		routersToConnections: make(map[*routerpkg.Router]server.Connection),
		connectionsToIDs:     make(map[server.Connection]*id.ID),
		idsToRouters:         make(map[*id.ID]*routerpkg.Router),
	}

	adapter.server.SetOnConnectedHandler(adapter.onConnectedHandler)

	return &adapter, nil
}

// Start begins the operation of the NetAdapter
func (na *NetAdapter) Start() error {
	err := na.server.Start()
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
	return na.server.Stop()
}

func (na *NetAdapter) Connect(address string) (server.Connection, error) {
	return na.server.Connect(address)
}

func (na *NetAdapter) Connections() []server.Connection {
	connections := make([]server.Connection, 0, len(na.connectionsToIDs))

	for connection := range na.connectionsToIDs {
		connections = append(connections, connection)
	}

	return connections
}

func (na *NetAdapter) onConnectedHandler(connection server.Connection) error {
	router, err := na.routerInitializer()
	if err != nil {
		return err
	}
	connection.Start(router)
	na.routersToConnections[router] = connection

	na.connectionsToIDs[connection] = nil

	router.SetOnRouteCapacityReachedHandler(func() {
		err := connection.Disconnect()
		if err != nil {
			log.Warnf("Failed to disconnect from %s", connection)
		}
	})
	connection.SetOnDisconnectedHandler(func() error {
		na.cleanupConnection(connection, router)
		na.server.RemoveConnection(connection)
		return router.Close()
	})
	na.server.AddConnection(connection)
	return nil
}

// AssociateRouterID associates the connection for the given router
// with the given ID
func (na *NetAdapter) AssociateRouterID(router *routerpkg.Router, id *id.ID) error {
	connection, ok := na.routersToConnections[router]
	if !ok {
		return errors.Errorf("router not registered for id %s", id)
	}

	na.connectionsToIDs[connection] = id
	na.idsToRouters[id] = router
	return nil
}

func (na *NetAdapter) cleanupConnection(connection server.Connection, router *routerpkg.Router) {
	connectionID, ok := na.connectionsToIDs[connection]
	if !ok {
		return
	}

	delete(na.routersToConnections, router)
	delete(na.connectionsToIDs, connection)
	delete(na.idsToRouters, connectionID)
}

// SetRouterInitializer sets the routerInitializer function
// for the net adapter
func (na *NetAdapter) SetRouterInitializer(routerInitializer RouterInitializer) {
	na.routerInitializer = routerInitializer
}

// ID returns this netAdapter's ID in the network
func (na *NetAdapter) ID() *id.ID {
	return na.id
}

// Broadcast sends the given `message` to every peer corresponding
// to each ID in `ids`
func (na *NetAdapter) Broadcast(connectionIDs []*id.ID, message wire.Message) error {
	na.RLock()
	defer na.RUnlock()
	for _, connectionID := range connectionIDs {
		router, ok := na.idsToRouters[connectionID]
		if !ok {
			log.Warnf("connectionID %s is not registered", connectionID)
			continue
		}
		_, err := router.EnqueueIncomingMessage(message)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetBestLocalAddress returns the most appropriate local address to use
// for the given remote address.
func (na *NetAdapter) GetBestLocalAddress() (*wire.NetAddress, error) {
	//TODO(libp2p) Reimplement this, and check reachability to the other node
	if len(config.ActiveConfig().ExternalIPs) > 0 {
		host, portString, err := net.SplitHostPort(config.ActiveConfig().ExternalIPs[0])
		if err != nil {
			portString = config.ActiveConfig().NetParams().DefaultPort
		}
		portInt, err := strconv.Atoi(portString)
		if err != nil {
			return nil, err
		}

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
		return wire.NewNetAddressIPPort(ip, uint16(portInt), wire.SFNodeNetwork), nil

	}
	listenAddress := config.ActiveConfig().Listeners[0]
	_, portString, err := net.SplitHostPort(listenAddress)
	if err != nil {
		portString = config.ActiveConfig().NetParams().DefaultPort
	}

	portInt, err := strconv.Atoi(portString)
	if err != nil {
		return nil, err
	}

	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, address := range addresses {
		ip, _, err := net.ParseCIDR(address.String())
		if err != nil {
			continue
		}

		return wire.NewNetAddressIPPort(ip, uint16(portInt), wire.SFNodeNetwork), nil
	}
	return nil, errors.New("no address was found")
}
