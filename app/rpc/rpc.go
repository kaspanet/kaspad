package rpc

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/app/rpc/rpchandlers"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
)

type handlerFunc func(context *rpccontext.Context, router *router.Router, request appmessage.Message) (appmessage.Message, error)

func defaultHandlerFunc(context *rpccontext.Context, router *router.Router, request appmessage.Message) (appmessage.Message, error) {
	return nil, errors.New("Not implemented")
}

var rpcHandlers = map[appmessage.MessageCommand]handlerFunc{
	appmessage.CmdNotifyBlockAddedRequestMessage:           defaultHandlerFunc,
	appmessage.CmdNotifyTransactionAddedRequestMessage:     defaultHandlerFunc,
	appmessage.CmdNotifyUTXOOfAddressChangedRequestMessage: defaultHandlerFunc,
	appmessage.CmdNotifyFinalityConflictsRequestMessage:    defaultHandlerFunc,
	appmessage.CmdNotifyChainChangedRequestMessage:         defaultHandlerFunc,
	appmessage.CmdGetCurrentNetworkRequestMessage:          rpchandlers.HandleGetCurrentNetwork,
	appmessage.CmdSubmitBlockRequestMessage:                rpchandlers.HandleSubmitBlock,
	appmessage.CmdGetBlockTemplateRequestMessage:           rpchandlers.HandleGetBlockTemplate,
	appmessage.CmdGetPeerAddressesRequestMessage:           rpchandlers.HandleGetPeerAddresses,
	appmessage.CmdGetSelectedTipHashRequestMessage:         rpchandlers.HandleGetSelectedTipHash,
	appmessage.CmdGetMempoolEntryRequestMessage:            rpchandlers.HandleGetMempoolEntry,
	appmessage.CmdGetConnectedPeerInfoRequestMessage:       rpchandlers.HandleGetConnectedPeerInfo,
	appmessage.CmdAddPeerRequestMessage:                    rpchandlers.HandleAddPeer,
	appmessage.CmdSubmitTransactionRequestMessage:          rpchandlers.HandleSubmitTransaction,
	appmessage.CmdGetBlockRequestMessage:                   rpchandlers.HandleGetBlock,
	appmessage.CmdGetSubnetworkRequestMessage:              rpchandlers.HandleGetSubnetwork,
	appmessage.CmdGetChainFromBlockRequestMessage:          rpchandlers.HandleGetChainFromBlock,
	appmessage.CmdGetBlocksRequestMessage:                  rpchandlers.HandleGetBlocks,
	appmessage.CmdGetBlockCountRequestMessage:              rpchandlers.HandleGetBlockCount,
	appmessage.CmdGetBlockDAGInfoRequestMessage:            rpchandlers.HandleGetBlockDAGInfo,
	appmessage.CmdResolveFinalityConflictRequestMessage:    rpchandlers.HandleResolveFinalityConflict,
	appmessage.CmdGetMempoolEntriesRequestMessage:          rpchandlers.HandleGetMempoolEntries,
	appmessage.CmdShutDownRequestMessage:                   rpchandlers.HandleShutDown,
	appmessage.CmdGetHeadersRequestMessage:                 rpchandlers.HandleGetHeaders,
	appmessage.CmdGetUTXOsByAddressRequestMessage:          rpchandlers.HandleGetUTXOsByAddress,
}

func (m *Manager) routerInitializer(router *router.Router, netConnection *netadapter.NetConnection) {
	messageTypes := make([]appmessage.MessageCommand, 0, len(m.handlers))
	for messageType := range m.handlers {
		messageTypes = append(messageTypes, messageType)
	}
	incomingRoute, err := router.AddIncomingRoute(messageTypes)
	if err != nil {
		panic(err)
	}
	m.context.NotificationManager.AddListener(router)

	spawn("routerInitializer-handleIncomingMessages", func() {
		defer m.context.NotificationManager.RemoveListener(router)

		err := m.handleIncomingMessages(router, incomingRoute)
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
		handler, ok := m.handlers[request.Command()]
		if !ok {
			return err
		}
		response, err := handler.Execute(router, request)
		if err != nil {
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

// RegisterHandler registers new rpc handler
func (m *Manager) RegisterHandler(command appmessage.MessageCommand, rpcHandler Handler) {
	m.handlers[command] = rpcHandler
}
