package rpc

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

type handler func(outgoingRoute *routerpkg.Route)

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
	spawn("routerInitializer", func() {
		handleIncomingMessages(incomingRoute, router.OutgoingRoute())
	})
}

func handleIncomingMessages(incomingRoute, outgoingRoute *routerpkg.Route) {
	for {
		message, err := incomingRoute.Dequeue()
		if err != nil {
			panic(err) // TODO
		}
		handler, ok := handlers[message.Command()]
		if !ok {
			panic(err) // TODO
		}
		handler(outgoingRoute)
	}
}

func handleGetCurrentNetwork(outgoingRoute *routerpkg.Route) {
	log.Warnf("GOT CURRENT NET REQUEST")
}
