package netadapter

import (
	"net"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/domainmessage"
	"github.com/kaspanet/kaspad/netadapter/id"
	routerpkg "github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/netadapter/server"
	"github.com/kaspanet/kaspad/netadapter/server/grpcserver"
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
	cfg               *config.Config
	id                *id.ID
	server            server.Server
	routerInitializer RouterInitializer
	stop              uint32

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
	s, err := grpcserver.NewGRPCServer(cfg.Listeners)
	if err != nil {
		return nil, err
	}
	adapter := NetAdapter{
		cfg:    cfg,
		id:     netAdapterID,
		server: s,

		connections: make(map[*NetConnection]struct{}),
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
	netConnection := newNetConnection(connection, na.routerInitializer)

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
// to each NetConnection in the given netConnections
func (na *NetAdapter) Broadcast(netConnections []*NetConnection, message domainmessage.Message) error {
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

// GetBestLocalAddress returns the most appropriate local address to use
// for the given remote address.
func (na *NetAdapter) GetBestLocalAddress() (*domainmessage.NetAddress, error) {
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
		return domainmessage.NewNetAddressIPPort(ip, uint16(portInt), domainmessage.SFNodeNetwork), nil

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

		return domainmessage.NewNetAddressIPPort(ip, uint16(portInt), domainmessage.SFNodeNetwork), nil
	}
	return nil, errors.New("no address was found")
}
