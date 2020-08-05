package netadaptermock

import (
	"sync"

	"github.com/kaspanet/kaspad/netadapter/id"
	"github.com/kaspanet/kaspad/util/mstime"

	"github.com/kaspanet/kaspad/protocol/common"

	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/wire"

	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/netadapter"

	"github.com/pkg/errors"
)

// NetAdapterMock allows tests and other tools to mockup a simple network adapter without implementing all the required
// supporting structures.
type NetAdapterMock struct {
	lock       sync.Mutex
	netAdapter *netadapter.NetAdapter
	routesChan <-chan *Routes
}

// New creates a new instance of a NetAdapterMock
func New(cfg *config.Config) (*NetAdapterMock, error) {
	netAdapter, err := netadapter.NewNetAdapter(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "Error starting netAdapter")
	}

	routerInitializer, routesChan := generateRouteInitializer()

	netAdapter.SetRouterInitializer(routerInitializer)
	err = netAdapter.Start()
	if err != nil {
		return nil, errors.Wrap(err, "Error starting netAdapter")
	}

	return &NetAdapterMock{
		lock:       sync.Mutex{},
		netAdapter: netAdapter,
		routesChan: routesChan,
	}, nil
}

// Connect opens a connection to the given address, handles handshake, and returns the routes for this connection
// To simplify usage the return type contains only two routes:
// OutgoingRoute - for all outgoing messages
// IncomingRoute - for all incoming messages (excluding handshake messages)
func (nam *NetAdapterMock) Connect(address string) (*Routes, error) {
	nam.lock.Lock()
	defer nam.lock.Unlock()

	err := nam.netAdapter.Connect(address)
	if err != nil {
		return nil, err
	}

	routes := <-nam.routesChan
	err = handleHandshake(routes, nam.netAdapter.ID())
	if err != nil {
		return nil, errors.Wrap(err, "Error in handshake")
	}

	spawn("netAdapterMock-handlePingPong", func() {
		err := handlePingPong(routes)
		if err != nil {
			panic(errors.Wrap(err, "Error from ping-pong"))
		}
	})

	return routes, nil
}

func handlePingPong(routes *Routes) error {
	for {
		message, err := routes.pingRoute.Dequeue()
		if err != nil {
			if errors.Is(err, router.ErrRouteClosed) {
				return nil
			}
			return err
		}

		pingMessage := message.(*wire.MsgPing)

		err = routes.OutgoingRoute.Enqueue(&wire.MsgPong{Nonce: pingMessage.Nonce})
		if err != nil {
			return err
		}
	}
}

func handleHandshake(routes *Routes, ourID *id.ID) error {
	msg, err := routes.handshakeRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return err
	}

	versionMessage, ok := msg.(*wire.MsgVersion)
	if !ok {
		return errors.Errorf("Expected first message to be of type %s, but got %s", wire.CmdVersion, msg.Command())
	}

	err = routes.OutgoingRoute.Enqueue(&wire.MsgVersion{
		ProtocolVersion: versionMessage.ProtocolVersion,
		Services:        versionMessage.Services,
		Timestamp:       mstime.Now(),
		Address:         nil,
		ID:              ourID,
		UserAgent:       "/net-adapter-mock/",
		SelectedTipHash: versionMessage.SelectedTipHash,
		DisableRelayTx:  true,
		SubnetworkID:    nil,
	})
	if err != nil {
		return err
	}

	msg, err = routes.handshakeRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return err
	}

	_, ok = msg.(*wire.MsgVerAck)
	if !ok {
		return errors.Errorf("Expected second message to be of type %s, but got %s", wire.CmdVerAck, msg.Command())
	}

	err = routes.OutgoingRoute.Enqueue(&wire.MsgVerAck{})
	if err != nil {
		return err
	}

	return nil
}

func generateRouteInitializer() (netadapter.RouterInitializer, <-chan *Routes) {
	everythingElse := make([]wire.MessageCommand, 0, len(wire.MessageCommandToString)-3)
	for command := range wire.MessageCommandToString {
		if command != wire.CmdVersion && command != wire.CmdVerAck && command != wire.CmdPing {
			everythingElse = append(everythingElse, command)
		}
	}

	routesChan := make(chan *Routes)

	routeInitializer := func(router *router.Router, netConnection *netadapter.NetConnection) {
		handshakeRoute, err := router.AddIncomingRoute([]wire.MessageCommand{wire.CmdVersion, wire.CmdVerAck})
		if err != nil {
			panic(errors.Wrap(err, "Error registering handshake route"))
		}
		pingRoute, err := router.AddIncomingRoute([]wire.MessageCommand{wire.CmdPing})
		if err != nil {
			panic(errors.Wrap(err, "Error registering ping route"))
		}

		everythingElseRoute, err := router.AddIncomingRoute(everythingElse)
		if err != nil {
			panic(errors.Wrap(err, "Error registering everythingElseRoute"))
		}

		spawn("netAdapterMock-routeInitializer-sendRoutesToChan", func() {
			routesChan <- &Routes{
				netConnection:  netConnection,
				OutgoingRoute:  router.OutgoingRoute(),
				IncomingRoute:  everythingElseRoute,
				handshakeRoute: handshakeRoute,
				pingRoute:      pingRoute,
			}
		})
	}

	return routeInitializer, routesChan
}
