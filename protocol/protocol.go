package protocol

import (
	"sync/atomic"

	"github.com/kaspanet/kaspad/protocol/flows/handshake"

	"github.com/kaspanet/kaspad/protocol/flows/addressexchange"
	"github.com/kaspanet/kaspad/protocol/flows/blockrelay"

	routerpkg "github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/flows/ping"
	"github.com/kaspanet/kaspad/protocol/flows/relaytransactions"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

func (m *Manager) routerInitializer() (*routerpkg.Router, error) {

	router := routerpkg.NewRouter()
	spawn(func() {
		err := m.startFlows(router)
		if err != nil {
			if protocolErr := &(protocolerrors.ProtocolError{}); errors.As(err, &protocolErr) {
				if protocolErr.ShouldBan {
					// TODO(libp2p) Ban peer
					panic("unimplemented")
				}
				err = m.netAdapter.DisconnectAssociatedConnection(router)
				if err != nil {
					panic(err)
				}
				return
			}
			if errors.Is(err, routerpkg.ErrTimeout) {
				err = m.netAdapter.DisconnectAssociatedConnection(router)
				if err != nil {
					panic(err)
				}
				return
			}
			panic(err)
		}
	})
	return router, nil

}

func (m *Manager) startFlows(router *routerpkg.Router) error {
	stop := make(chan error)
	stopped := uint32(0)

	outgoingRoute := router.OutgoingRoute()
	peer := new(peerpkg.Peer)

	closed, err := handshake.HandleHandshake(router, m.netAdapter, peer, m.dag, m.addressManager)
	if err != nil {
		return err
	}
	if closed {
		return nil
	}

	addOneTimeFlow("SendAddresses", router, []wire.MessageCommand{wire.CmdGetAddresses}, &stopped, stop,
		func(incomingRoute *routerpkg.Route) (routeClosed bool, err error) {
			return addressexchange.SendAddresses(incomingRoute, outgoingRoute, m.addressManager)
		},
	)

	addOneTimeFlow("ReceiveAddresses", router, []wire.MessageCommand{wire.CmdAddress}, &stopped, stop,
		func(incomingRoute *routerpkg.Route) (routeClosed bool, err error) {
			return addressexchange.ReceiveAddresses(incomingRoute, outgoingRoute, peer, m.addressManager)
		},
	)

	addFlow("HandleRelayInvs", router, []wire.MessageCommand{wire.CmdInvRelayBlock, wire.CmdBlock}, &stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return blockrelay.HandleRelayInvs(incomingRoute,
				outgoingRoute, peer, m.netAdapter, m.dag, m.OnNewBlock)
		},
	)

	addFlow("HandleRelayBlockRequests", router, []wire.MessageCommand{wire.CmdGetRelayBlocks}, &stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return blockrelay.HandleRelayBlockRequests(incomingRoute, outgoingRoute, peer, m.dag)
		},
	)

	addFlow("ReceivePings", router, []wire.MessageCommand{wire.CmdPing}, &stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ping.ReceivePings(incomingRoute, outgoingRoute)
		},
	)

	addFlow("SendPings", router, []wire.MessageCommand{wire.CmdPong}, &stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ping.SendPings(incomingRoute, outgoingRoute, peer)
		},
	)

	addFlow("RelayedTransactions", router, []wire.MessageCommand{wire.CmdInv, wire.CmdTx}, &stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return relaytransactions.HandleRelayedTransactions(incomingRoute, outgoingRoute, m.netAdapter, m.dag,
				m.txPool, m.sharedRequestedTransactions)
		},
	)

	err = <-stop
	return err
}

func addFlow(name string, router *routerpkg.Router, messageTypes []wire.MessageCommand, stopped *uint32,
	stopChan chan error, flow func(route *routerpkg.Route) error) {

	route, err := router.AddIncomingRoute(messageTypes)
	if err != nil {
		panic(err)
	}

	spawn(func() {
		err := flow(route)
		if err != nil {
			log.Errorf("error from %s flow: %s", name, err)
		}
		if atomic.AddUint32(stopped, 1) == 1 {
			stopChan <- err
		}
	})
}

func addOneTimeFlow(name string, router *routerpkg.Router, messageTypes []wire.MessageCommand, stopped *uint32,
	stopChan chan error, flow func(route *routerpkg.Route) (routeClosed bool, err error)) {

	route, err := router.AddIncomingRoute(messageTypes)
	if err != nil {
		panic(err)
	}

	spawn(func() {
		defer func() {
			err := router.RemoveRoute(messageTypes)
			if err != nil {
				panic(err)
			}
		}()

		closed, err := flow(route)
		if err != nil {
			log.Errorf("error from %s flow: %s", name, err)
		}
		if (err != nil || closed) && atomic.AddUint32(stopped, 1) == 1 {
			stopChan <- err
		}
	})
}
