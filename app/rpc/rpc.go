package rpc

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/app/rpc/rpcerrors"
	"github.com/kaspanet/kaspad/app/rpc/rpchandlers"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
)

type handler func(context *rpccontext.Context, request appmessage.Message) (appmessage.Message, error)

var handlers = map[appmessage.MessageCommand]handler{
	appmessage.CmdGetCurrentNetworkRequestMessage: rpchandlers.HandleGetCurrentNetwork,
	appmessage.CmdSubmitBlockRequestMessage:       rpchandlers.HandleSubmitBlock,
	appmessage.CmdGetBlockTemplateRequestMessage:  rpchandlers.HandleGetBlockTemplate,
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
		err := m.handleIncomingMessages(incomingRoute, router.OutgoingRoute())
		m.handleError(err, netConnection)
	})
}

func (m *Manager) handleIncomingMessages(incomingRoute, outgoingRoute *router.Route) error {
	for {
		request, err := incomingRoute.Dequeue()
		if err != nil {
			return err
		}
		handler, ok := handlers[request.Command()]
		if !ok {
			return err
		}
		response, err := handler(m.context, request)
		if err != nil {
			if rpcErr := &(rpcerrors.RPCError{}); errors.As(err, &rpcErr) {
				errorMessage := appmessage.NewRPCErrorMessage(rpcErr.Message)
				return outgoingRoute.Enqueue(errorMessage)
			}
			return err
		}
		err = outgoingRoute.Enqueue(response)
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
