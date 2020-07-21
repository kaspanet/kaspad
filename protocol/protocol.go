package protocol

import (
	"fmt"
	"sync/atomic"

	"github.com/kaspanet/kaspad/protocol/flows/handshake"

	"github.com/kaspanet/kaspad/protocol/flows/addressexchange"
	"github.com/kaspanet/kaspad/protocol/flows/blockrelay"

	routerpkg "github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/flows/ibd"
	"github.com/kaspanet/kaspad/protocol/flows/ping"
	"github.com/kaspanet/kaspad/protocol/flows/relaytransactions"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

func (m *Manager) routerInitializer() (*routerpkg.Router, error) {

	router := routerpkg.NewRouter()
	spawn("newRouterInitializer-startFlows", func() {
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
			if errors.Is(err, routerpkg.ErrRouteClosed) {
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

	peer, closed, err := handshake.HandleHandshake(m.cfg, router, m.netAdapter, m.dag, m.addressManager)
	if err != nil {
		return err
	}
	if closed {
		return nil
	}

	m.addAddressFlows(router, &stopped, stop, peer)
	m.addBlockRelayFlows(router, &stopped, stop, peer)
	m.addPingFlows(router, &stopped, stop, peer)
	m.addIBDFlows(router, &stopped, stop, peer)
	m.addTransactionRelayFlow(router, &stopped, stop)

	err = <-stop
	return err
}

func (m *Manager) addAddressFlows(router *routerpkg.Router, stopped *uint32, stop chan error,
	peer *peerpkg.Peer) {

	outgoingRoute := router.OutgoingRoute()

	addOneTimeFlow("SendAddresses", router, []wire.MessageCommand{wire.CmdGetAddresses}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return addressexchange.SendAddresses(incomingRoute, outgoingRoute, m.addressManager)
		},
	)

	addOneTimeFlow("ReceiveAddresses", router, []wire.MessageCommand{wire.CmdAddress}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return addressexchange.ReceiveAddresses(incomingRoute, outgoingRoute, m.cfg, peer, m.addressManager)
		},
	)
}

func (m *Manager) addBlockRelayFlows(router *routerpkg.Router, stopped *uint32, stop chan error, peer *peerpkg.Peer) {

	outgoingRoute := router.OutgoingRoute()

	addFlow("HandleRelayInvs", router, []wire.MessageCommand{wire.CmdInvRelayBlock, wire.CmdBlock}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return blockrelay.HandleRelayInvs(incomingRoute,
				outgoingRoute, peer, m.netAdapter, m.dag, m.OnNewBlock)
		},
	)

	addFlow("HandleRelayBlockRequests", router, []wire.MessageCommand{wire.CmdGetRelayBlocks}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return blockrelay.HandleRelayBlockRequests(incomingRoute, outgoingRoute, peer, m.dag)
		},
	)
}

func (m *Manager) addPingFlows(router *routerpkg.Router, stopped *uint32, stop chan error, peer *peerpkg.Peer) {
	outgoingRoute := router.OutgoingRoute()

	addFlow("ReceivePings", router, []wire.MessageCommand{wire.CmdPing}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ping.ReceivePings(incomingRoute, outgoingRoute)
		},
	)

	addFlow("SendPings", router, []wire.MessageCommand{wire.CmdPong}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ping.SendPings(incomingRoute, outgoingRoute, peer)
		},
	)
}

func (m *Manager) addIBDFlows(router *routerpkg.Router, stopped *uint32, stop chan error,
	peer *peerpkg.Peer) {

	outgoingRoute := router.OutgoingRoute()

	addFlow("HandleIBD", router, []wire.MessageCommand{wire.CmdBlockLocator, wire.CmdIBDBlock}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ibd.HandleIBD(incomingRoute, outgoingRoute, peer, m.dag, m.OnNewBlock)
		},
	)

	addFlow("RequestSelectedTip", router, []wire.MessageCommand{wire.CmdSelectedTip}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ibd.RequestSelectedTip(incomingRoute, outgoingRoute, peer, m.dag)
		},
	)

	addFlow("HandleGetSelectedTip", router, []wire.MessageCommand{wire.CmdGetSelectedTip}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ibd.HandleGetSelectedTip(incomingRoute, outgoingRoute, m.dag)
		},
	)

	addFlow("HandleGetBlockLocator", router, []wire.MessageCommand{wire.CmdGetBlockLocator}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ibd.HandleGetBlockLocator(incomingRoute, outgoingRoute, m.dag)
		},
	)

	addFlow("HandleGetBlocks", router, []wire.MessageCommand{wire.CmdGetBlocks}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ibd.HandleGetBlocks(incomingRoute, outgoingRoute, m.dag)
		},
	)
}

func (m *Manager) addTransactionRelayFlow(router *routerpkg.Router, stopped *uint32, stop chan error) {

	outgoingRoute := router.OutgoingRoute()

	addFlow("HandleRelayedTransactions", router, []wire.MessageCommand{wire.CmdInv, wire.CmdTx}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return relaytransactions.HandleRelayedTransactions(incomingRoute, outgoingRoute, m.netAdapter, m.dag,
				m.txPool, m.sharedRequestedTransactions)
		},
	)
}

func addFlow(name string, router *routerpkg.Router, messageTypes []wire.MessageCommand, stopped *uint32,
	stopChan chan error, flow func(route *routerpkg.Route) error) {

	route, err := router.AddIncomingRoute(messageTypes)
	if err != nil {
		panic(err)
	}

	spawn(fmt.Sprintf("addFlow-startFlow-%s", name), func() {
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
	stopChan chan error, flow func(route *routerpkg.Route) error) {

	route, err := router.AddIncomingRoute(messageTypes)
	if err != nil {
		panic(err)
	}

	spawn(fmt.Sprintf("addOneTimeFlow-startFlow-%s", name), func() {
		defer func() {
			err := router.RemoveRoute(messageTypes)
			if err != nil {
				panic(err)
			}
		}()

		err := flow(route)
		if err != nil {
			log.Errorf("error from %s flow: %s", name, err)
		}
		if err != nil && atomic.AddUint32(stopped, 1) == 1 {
			stopChan <- err
		}
	})
}
