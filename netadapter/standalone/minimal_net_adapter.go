package standalone

import (
	"github.com/kaspanet/kaspad/util/mstime"
	"sync"

	"github.com/kaspanet/kaspad/netadapter/id"
	"github.com/kaspanet/kaspad/protocol/common"

	"github.com/kaspanet/kaspad/domainmessage"
	"github.com/kaspanet/kaspad/netadapter/router"

	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/netadapter"

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

	netAdapter.SetRouterInitializer(routerInitializer)
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

	err := mna.netAdapter.Connect(address)
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

		pingMessage := message.(*domainmessage.MsgPing)

		err = routes.OutgoingRoute.Enqueue(&domainmessage.MsgPong{Nonce: pingMessage.Nonce})
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

	versionMessage, ok := msg.(*domainmessage.MsgVersion)
	if !ok {
		return errors.Errorf("expected first message to be of type %s, but got %s", domainmessage.CmdVersion, msg.Command())
	}

	err = routes.OutgoingRoute.Enqueue(&domainmessage.MsgVersion{
		ProtocolVersion: versionMessage.ProtocolVersion,
		Network:         mna.cfg.ActiveNetParams.Name,
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

	_, ok = msg.(*domainmessage.MsgVerAck)
	if !ok {
		return errors.Errorf("expected second message to be of type %s, but got %s", domainmessage.CmdVerAck, msg.Command())
	}

	err = routes.OutgoingRoute.Enqueue(&domainmessage.MsgVerAck{})
	if err != nil {
		return err
	}

	return nil
}

func generateRouteInitializer() (netadapter.RouterInitializer, <-chan *Routes) {
	cmdsWithBuiltInRoutes := []domainmessage.MessageCommand{domainmessage.CmdVerAck, domainmessage.CmdVersion, domainmessage.CmdPing}

	everythingElse := make([]domainmessage.MessageCommand, 0, len(domainmessage.MessageCommandToString)-len(cmdsWithBuiltInRoutes))
outerLoop:
	for command := range domainmessage.MessageCommandToString {
		for _, cmdWithBuiltInRoute := range cmdsWithBuiltInRoutes {
			if command == cmdWithBuiltInRoute {
				continue outerLoop
			}
		}

		everythingElse = append(everythingElse, command)
	}

	routesChan := make(chan *Routes)

	routeInitializer := func(router *router.Router, netConnection *netadapter.NetConnection) {
		handshakeRoute, err := router.AddIncomingRoute([]domainmessage.MessageCommand{domainmessage.CmdVersion, domainmessage.CmdVerAck})
		if err != nil {
			panic(errors.Wrap(err, "error registering handshake route"))
		}
		pingRoute, err := router.AddIncomingRoute([]domainmessage.MessageCommand{domainmessage.CmdPing})
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
				pingRoute:      pingRoute,
			}
		})
	}

	return routeInitializer, routesChan
}
