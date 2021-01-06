package standalone

import (
	"sync"

	"github.com/kaspanet/kaspad/app/protocol/common"
	"github.com/kaspanet/kaspad/util/mstime"

	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/id"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"

	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"

	"github.com/pkg/errors"
)

// MinimalNetAdapter allows tests and other tools to use a simple network adapter without implementing
// all the required supporting structures.
type MinimalNetAdapter struct {
	cfg        *config.Config
	lock       sync.Mutex
	netAdapter *netadapter.NetAdapter
	routesChan <-chan *Routes
}

// NewMinimalNetAdapter creates a new instance of a MinimalNetAdapter
func NewMinimalNetAdapter(cfg *config.Config) (*MinimalNetAdapter, error) {
	netAdapter, err := netadapter.NewNetAdapter(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "Error starting netAdapter")
	}

	routerInitializer, routesChan := generateRouteInitializer()

	netAdapter.SetP2PRouterInitializer(routerInitializer)
	netAdapter.SetRPCRouterInitializer(func(_ *router.Router, _ *netadapter.NetConnection) {
	})

	err = netAdapter.Start()
	if err != nil {
		return nil, errors.Wrap(err, "Error starting netAdapter")
	}

	return &MinimalNetAdapter{
		cfg:        cfg,
		lock:       sync.Mutex{},
		netAdapter: netAdapter,
		routesChan: routesChan,
	}, nil
}

// Connect opens a connection to the given address, handles handshake, and returns the routes for this connection
// To simplify usage the return type contains only two routes:
// OutgoingRoute - for all outgoing messages
// IncomingRoute - for all incoming messages (excluding handshake messages)
func (mna *MinimalNetAdapter) Connect(address string) (*Routes, error) {
	mna.lock.Lock()
	defer mna.lock.Unlock()

	err := mna.netAdapter.P2PConnect(address)
	if err != nil {
		return nil, err
	}

	routes := <-mna.routesChan
	err = mna.handleHandshake(routes, mna.netAdapter.ID())
	if err != nil {
		return nil, errors.Wrap(err, "Error in handshake")
	}

	spawn("netAdapterMock-handlePingPong", func() {
		err := mna.handlePingPong(routes)
		if err != nil {
			panic(errors.Wrap(err, "Error from ping-pong"))
		}
	})

	return routes, nil
}

// handlePingPong makes sure that we are not disconnected due to not responding to pings.
// However, it only responds to pings, not sending its own, to conform to the minimal-ness
// of MinimalNetAdapter
func (*MinimalNetAdapter) handlePingPong(routes *Routes) error {
	for {
		message, err := routes.pingRoute.Dequeue()
		if err != nil {
			if errors.Is(err, router.ErrRouteClosed) {
				return nil
			}
			return err
		}

		pingMessage := message.(*appmessage.MsgPing)

		err = routes.OutgoingRoute.Enqueue(&appmessage.MsgPong{Nonce: pingMessage.Nonce})
		if err != nil {
			return err
		}
	}
}

func (mna *MinimalNetAdapter) handleHandshake(routes *Routes, ourID *id.ID) error {
	msg, err := routes.handshakeRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return err
	}
	versionMessage, ok := msg.(*appmessage.MsgVersion)
	if !ok {
		return errors.Errorf("expected first message to be of type %s, but got %s", appmessage.CmdVersion, msg.Command())
	}
	err = routes.OutgoingRoute.Enqueue(&appmessage.MsgVersion{
		ProtocolVersion: versionMessage.ProtocolVersion,
		Network:         mna.cfg.ActiveNetParams.Name,
		Services:        versionMessage.Services,
		Timestamp:       mstime.Now(),
		Address:         nil,
		ID:              ourID,
		UserAgent:       "/net-adapter-mock/",
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
	_, ok = msg.(*appmessage.MsgVerAck)
	if !ok {
		return errors.Errorf("expected second message to be of type %s, but got %s", appmessage.CmdVerAck, msg.Command())
	}
	err = routes.OutgoingRoute.Enqueue(&appmessage.MsgVerAck{})
	if err != nil {
		return err
	}

	msg, err = routes.addressesRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return err
	}
	_, ok = msg.(*appmessage.MsgRequestAddresses)
	if !ok {
		return errors.Errorf("expected third message to be of type %s, but got %s", appmessage.CmdRequestAddresses, msg.Command())
	}
	err = routes.OutgoingRoute.Enqueue(&appmessage.MsgAddresses{
		AddressList: []*appmessage.NetAddress{},
	})
	if err != nil {
		return err
	}

	err = routes.OutgoingRoute.Enqueue(&appmessage.MsgRequestAddresses{
		IncludeAllSubnetworks: true,
		SubnetworkID:          nil,
	})
	if err != nil {
		return err
	}
	msg, err = routes.addressesRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return err
	}
	_, ok = msg.(*appmessage.MsgAddresses)
	if !ok {
		return errors.Errorf("expected fourth message to be of type %s, but got %s", appmessage.CmdAddresses, msg.Command())
	}

	return nil
}

func generateRouteInitializer() (netadapter.RouterInitializer, <-chan *Routes) {
	cmdsWithBuiltInRoutes := []appmessage.MessageCommand{
		appmessage.CmdVersion,
		appmessage.CmdVerAck,
		appmessage.CmdRequestAddresses,
		appmessage.CmdAddresses,
		appmessage.CmdPing}

	everythingElse := make([]appmessage.MessageCommand, 0, len(appmessage.ProtocolMessageCommandToString)-len(cmdsWithBuiltInRoutes))
outerLoop:
	for command := range appmessage.ProtocolMessageCommandToString {
		for _, cmdWithBuiltInRoute := range cmdsWithBuiltInRoutes {
			if command == cmdWithBuiltInRoute {
				continue outerLoop
			}
		}

		everythingElse = append(everythingElse, command)
	}

	routesChan := make(chan *Routes)

	routeInitializer := func(router *router.Router, netConnection *netadapter.NetConnection) {
		handshakeRoute, err := router.AddIncomingRoute([]appmessage.MessageCommand{appmessage.CmdVersion, appmessage.CmdVerAck})
		if err != nil {
			panic(errors.Wrap(err, "error registering handshake route"))
		}
		addressesRoute, err := router.AddIncomingRoute([]appmessage.MessageCommand{appmessage.CmdRequestAddresses, appmessage.CmdAddresses})
		if err != nil {
			panic(errors.Wrap(err, "error registering addresses route"))
		}
		pingRoute, err := router.AddIncomingRoute([]appmessage.MessageCommand{appmessage.CmdPing})
		if err != nil {
			panic(errors.Wrap(err, "error registering ping route"))
		}
		everythingElseRoute, err := router.AddIncomingRoute(everythingElse)
		if err != nil {
			panic(errors.Wrap(err, "error registering everythingElseRoute"))
		}

		spawn("netAdapterMock-routeInitializer-sendRoutesToChan", func() {
			routesChan <- &Routes{
				netConnection:  netConnection,
				OutgoingRoute:  router.OutgoingRoute(),
				IncomingRoute:  everythingElseRoute,
				handshakeRoute: handshakeRoute,
				addressesRoute: addressesRoute,
				pingRoute:      pingRoute,
			}
		})
	}

	return routeInitializer, routesChan
}
