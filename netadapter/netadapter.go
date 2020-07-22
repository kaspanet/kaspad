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
type RouterInitializer func(netConnection *NetConnection) (*routerpkg.Router, error)

// NetAdapter is an abstraction layer over networking.
// This type expects a RouteInitializer function. This
// function weaves together the various "routes" (messages
// and message handlers) without exposing anything related
// to networking internals.
type NetAdapter struct {
	cfg               *config.Config
	id                *id.ID
	server            server.Server
	routerInitializer RouterInitializer
	stop              uint32

	routersToConnections map[*routerpkg.Router]*NetConnection
	connectionsToIDs     map[*NetConnection]*id.ID
	idsToRouters         map[*id.ID]*routerpkg.Router
	sync.RWMutex
}

// NewNetAdapter creates and starts a new NetAdapter on the
// given listeningPort
func NewNetAdapter(cfg *config.Config) (*NetAdapter, error) {
	netAdapterID, err := id.GenerateID()
	if err != nil {
		return nil, err
	}
	s, err := grpcserver.NewGRPCServer(cfg.Listeners)
	if err != nil {
		return nil, err
	}
	adapter := NetAdapter{
		cfg:    cfg,
		id:     netAdapterID,
		server: s,

		routersToConnections: make(map[*routerpkg.Router]*NetConnection),
		connectionsToIDs:     make(map[*NetConnection]*id.ID),
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

// Connect tells the NetAdapter's underlying server to initiate a connection
// to the given address
func (na *NetAdapter) Connect(address string) error {
	_, err := na.server.Connect(address)
	return err
}

// Connections returns a list of connections currently connected and active
func (na *NetAdapter) Connections() []*NetConnection {
	netConnections := make([]*NetConnection, 0, len(na.connectionsToIDs))

	for netConnection := range na.connectionsToIDs {
		netConnections = append(netConnections, netConnection)
	}

	return netConnections
}

// ConnectionCount returns the count of the connected connections
func (na *NetAdapter) ConnectionCount() int {
	return len(na.connectionsToIDs)
}

func (na *NetAdapter) onConnectedHandler(connection server.Connection) error {
	netConnection := newNetConnection(connection, nil)
	router, err := na.routerInitializer(netConnection)
	if err != nil {
		return err
	}
	connection.Start(router)

	na.routersToConnections[router] = netConnection

	na.connectionsToIDs[netConnection] = nil

	router.SetOnRouteCapacityReachedHandler(func() {
		err := connection.Disconnect()
		if err != nil {
			if !errors.Is(err, server.ErrNetwork) {
				panic(err)
			}
			log.Warnf("Failed to disconnect from %s", connection)
		}
	})
	connection.SetOnDisconnectedHandler(func() error {
		na.cleanupConnection(netConnection, router)
		return router.Close()
	})
	return nil
}

// AssociateRouterID associates the connection for the given router
// with the given ID
func (na *NetAdapter) AssociateRouterID(router *routerpkg.Router, id *id.ID) error {
	netConnection, ok := na.routersToConnections[router]
	if !ok {
		return errors.Errorf("router not registered for id %s", id)
	}

	netConnection.id = id

	na.connectionsToIDs[netConnection] = id
	na.idsToRouters[id] = router
	return nil
}

func (na *NetAdapter) cleanupConnection(netConnection *NetConnection, router *routerpkg.Router) {
	connectionID, ok := na.connectionsToIDs[netConnection]
	if !ok {
		return
	}

	delete(na.routersToConnections, router)
	delete(na.connectionsToIDs, netConnection)
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
		err := router.EnqueueIncomingMessage(message)
		if err != nil {
			if errors.Is(err, routerpkg.ErrRouteClosed) {
				connection := na.routersToConnections[router]
				log.Debugf("Cannot enqueue message to %s: router is closed", connection)
				continue
			}
			return err
		}
	}
	return nil
}

// GetBestLocalAddress returns the most appropriate local address to use
// for the given remote address.
func (na *NetAdapter) GetBestLocalAddress() (*wire.NetAddress, error) {
	//TODO(libp2p) Reimplement this, and check reachability to the other node
	if len(na.cfg.ExternalIPs) > 0 {
		host, portString, err := net.SplitHostPort(na.cfg.ExternalIPs[0])
		if err != nil {
			portString = na.cfg.NetParams().DefaultPort
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
	listenAddress := na.cfg.Listeners[0]
	_, portString, err := net.SplitHostPort(listenAddress)
	if err != nil {
		portString = na.cfg.NetParams().DefaultPort
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

// DisconnectAssociatedConnection disconnects from the connection associated with the given router.
func (na *NetAdapter) DisconnectAssociatedConnection(router *routerpkg.Router) error {
	netConnection := na.routersToConnections[router]
	return na.Disconnect(netConnection)
}

// Disconnect disconnects the given connection
func (na *NetAdapter) Disconnect(netConnection *NetConnection) error {
	err := netConnection.connection.Disconnect()
	if err != nil {
		if !errors.Is(err, server.ErrNetwork) {
			return err
		}
		log.Warnf("Error disconnecting from %s: %s", netConnection, err)
	}
	return nil
}
