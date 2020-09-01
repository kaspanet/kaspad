package rpc

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/app/rpc/rpchandlers"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
)

type handler func(context *rpccontext.Context, router *router.Router, request appmessage.Message) (appmessage.Message, error)

var handlers = map[appmessage.MessageCommand]handler{
	appmessage.CmdGetCurrentNetworkRequestMessage:    rpchandlers.HandleGetCurrentNetwork,
	appmessage.CmdSubmitBlockRequestMessage:          rpchandlers.HandleSubmitBlock,
	appmessage.CmdGetBlockTemplateRequestMessage:     rpchandlers.HandleGetBlockTemplate,
	appmessage.CmdNotifyBlockAddedRequestMessage:     rpchandlers.HandleNotifyBlockAdded,
	appmessage.CmdGetPeerAddressesRequestMessage:     rpchandlers.HandleGetPeerAddresses,
	appmessage.CmdGetSelectedTipHashRequestMessage:   rpchandlers.HandleGetSelectedTipHash,
	appmessage.CmdGetMempoolEntryRequestMessage:      rpchandlers.HandleGetMempoolEntry,
	appmessage.CmdGetConnectedPeerInfoRequestMessage: rpchandlers.HandleGetConnectedPeerInfo,
}

func (m *Manager) routerInitializer(router *router.Router, netConnection *netadapter.NetConnection) {
	messageTypes := make([]appmessage.MessageCommand, 0, len(handlers))
	for messageType := range handlers {
		messageTypes = append(messageTypes, messageType)
	}
	incomingRoute, err := router.AddIncomingRoute(messageTypes)
	if err != nil {
		panic(err)
	}
	spawn("routerInitializer-handleIncomingMessages", func() {
		err := m.handleIncomingMessages(router, incomingRoute)
		m.handleError(err, netConnection)
	})
	spawn("routerInitializer-handleOutgoingNotifications", func() {
		err := m.handleOutgoingNotifications(router)
		m.handleError(err, netConnection)
	})
}

func (m *Manager) handleIncomingMessages(router *router.Router, incomingRoute *router.Route) error {
	outgoingRoute := router.OutgoingRoute()
	for {
		request, err := incomingRoute.Dequeue()
		if err != nil {
			return err
		}
		handler, ok := handlers[request.Command()]
		if !ok {
			return err
		}
		response, err := handler(m.context, router, request)
		if err != nil {
			return err
		}
		err = outgoingRoute.Enqueue(response)
		if err != nil {
			return err
		}
	}
}

func (m *Manager) handleOutgoingNotifications(router *router.Router) error {
	notificationListener := m.context.NotificationManager.AddListener(router)
	defer m.context.NotificationManager.RemoveListener(router)
	for {
		err := notificationListener.ProcessNextNotification()
		if err != nil {
			return err
		}
	}
}

func (m *Manager) handleError(err error, netConnection *netadapter.NetConnection) {
	if errors.Is(err, router.ErrTimeout) {
		log.Warnf("Got timeout from %s. Disconnecting...", netConnection)
		netConnection.Disconnect()
		return
	}
	if errors.Is(err, router.ErrRouteClosed) {
		return
	}
	panic(err)
}
