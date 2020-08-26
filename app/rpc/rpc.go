package rpc

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/app/rpc/rpchandlers"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

type handler func(context *rpccontext.Context, outgoingRoute *router.Route) error

var handlers = map[appmessage.MessageCommand]handler{
	appmessage.CmdGetCurrentNetworkRequestMessage: rpchandlers.HandleGetCurrentNetwork,
}

func (m *Manager) routerInitializer(router *router.Router, _ *netadapter.NetConnection) {
	messageTypes := make([]appmessage.MessageCommand, 0, len(handlers))
	for messageType := range handlers {
		messageTypes = append(messageTypes, messageType)
	}
	incomingRoute, err := router.AddIncomingRoute(messageTypes)
	if err != nil {
		log.Warnf("a %s", err) // TODO
		return
	}
	spawn("routerInitializer-handleIncomingMessages", func() {
		m.handleIncomingMessages(incomingRoute, router.OutgoingRoute())
	})
}

func (m *Manager) handleIncomingMessages(incomingRoute, outgoingRoute *router.Route) {
	for {
		message, err := incomingRoute.Dequeue()
		if err != nil {
			log.Warnf("a %s", err) // TODO
			return
		}
		handler, ok := handlers[message.Command()]
		if !ok {
			log.Warnf("a %s", err) // TODO
			return
		}
		err = handler(m.context, outgoingRoute)
		if err != nil {
			log.Warnf("a %s", err) // TODO
			return
		}
	}
}
