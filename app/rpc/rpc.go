package rpc

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

type handler func(context *context, outgoingRoute *routerpkg.Route) error

var handlers = map[appmessage.MessageCommand]handler{
	appmessage.CmdGetCurrentNetworkRequestMessage: handleGetCurrentNetwork,
}

func (m *Manager) routerInitializer(router *routerpkg.Router, netConnection *netadapter.NetConnection) {
	messageTypes := make([]appmessage.MessageCommand, 0, len(handlers))
	for messageType := range handlers {
		messageTypes = append(messageTypes, messageType)
	}
	incomingRoute, err := router.AddIncomingRoute(messageTypes)
	if err != nil {
		panic(err) // TODO
	}
	spawn("routerInitializer-handleIncomingMessages", func() {
		m.handleIncomingMessages(incomingRoute, router.OutgoingRoute())
	})
}

func (m *Manager) handleIncomingMessages(incomingRoute, outgoingRoute *routerpkg.Route) {
	for {
		message, err := incomingRoute.Dequeue()
		if err != nil {
			panic(err) // TODO
		}
		handler, ok := handlers[message.Command()]
		if !ok {
			panic(err) // TODO
		}
		err = handler(m.context, outgoingRoute)
		if err != nil {
			panic(err) // TODO
		}
	}
}

func handleGetCurrentNetwork(context *context, outgoingRoute *routerpkg.Route) error {
	log.Warnf("GOT CURRENT NET REQUEST")
	log.Warnf("HERE'S THE CURRENT NET: %s", context.dag.Params.Name)

	message := appmessage.NewGetCurrentVersionResponseMessage(context.dag.Params.Name)
	return outgoingRoute.Enqueue(message)
}
