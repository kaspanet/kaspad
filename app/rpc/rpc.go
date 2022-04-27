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
	appmessage.CmdGetCurrentNetworkRequestMessage:                           rpchandlers.HandleGetCurrentNetwork,
	appmessage.CmdSubmitBlockRequestMessage:                                 rpchandlers.HandleSubmitBlock,
	appmessage.CmdGetBlockTemplateRequestMessage:                            rpchandlers.HandleGetBlockTemplate,
	appmessage.CmdNotifyBlockAddedRequestMessage:                            rpchandlers.HandleNotifyBlockAdded,
	appmessage.CmdGetPeerAddressesRequestMessage:                            rpchandlers.HandleGetPeerAddresses,
	appmessage.CmdGetSelectedTipHashRequestMessage:                          rpchandlers.HandleGetSelectedTipHash,
	appmessage.CmdGetMempoolEntryRequestMessage:                             rpchandlers.HandleGetMempoolEntry,
	appmessage.CmdGetConnectedPeerInfoRequestMessage:                        rpchandlers.HandleGetConnectedPeerInfo,
	appmessage.CmdAddPeerRequestMessage:                                     rpchandlers.HandleAddPeer,
	appmessage.CmdSubmitTransactionRequestMessage:                           rpchandlers.HandleSubmitTransaction,
	appmessage.CmdNotifyVirtualSelectedParentChainChangedRequestMessage:     rpchandlers.HandleNotifyVirtualSelectedParentChainChanged,
	appmessage.CmdGetBlockRequestMessage:                                    rpchandlers.HandleGetBlock,
	appmessage.CmdGetSubnetworkRequestMessage:                               rpchandlers.HandleGetSubnetwork,
	appmessage.CmdGetVirtualSelectedParentChainFromBlockRequestMessage:      rpchandlers.HandleGetVirtualSelectedParentChainFromBlock,
	appmessage.CmdGetBlocksRequestMessage:                                   rpchandlers.HandleGetBlocks,
	appmessage.CmdGetBlockCountRequestMessage:                               rpchandlers.HandleGetBlockCount,
	appmessage.CmdGetBalanceByAddressRequestMessage:                         rpchandlers.HandleGetBalanceByAddress,
	appmessage.CmdGetBlockDAGInfoRequestMessage:                             rpchandlers.HandleGetBlockDAGInfo,
	appmessage.CmdResolveFinalityConflictRequestMessage:                     rpchandlers.HandleResolveFinalityConflict,
	appmessage.CmdNotifyFinalityConflictsRequestMessage:                     rpchandlers.HandleNotifyFinalityConflicts,
	appmessage.CmdGetMempoolEntriesRequestMessage:                           rpchandlers.HandleGetMempoolEntries,
	appmessage.CmdShutDownRequestMessage:                                    rpchandlers.HandleShutDown,
	appmessage.CmdGetHeadersRequestMessage:                                  rpchandlers.HandleGetHeaders,
	appmessage.CmdNotifyUTXOsChangedRequestMessage:                          rpchandlers.HandleNotifyUTXOsChanged,
	appmessage.CmdStopNotifyingUTXOsChangedRequestMessage:                   rpchandlers.HandleStopNotifyingUTXOsChanged,
	appmessage.CmdGetUTXOsByAddressesRequestMessage:                         rpchandlers.HandleGetUTXOsByAddresses,
	appmessage.CmdGetBalancesByAddressesRequestMessage:                      rpchandlers.HandleGetBalancesByAddresses,
	appmessage.CmdGetVirtualSelectedParentBlueScoreRequestMessage:           rpchandlers.HandleGetVirtualSelectedParentBlueScore,
	appmessage.CmdNotifyVirtualSelectedParentBlueScoreChangedRequestMessage: rpchandlers.HandleNotifyVirtualSelectedParentBlueScoreChanged,
	appmessage.CmdBanRequestMessage:                                         rpchandlers.HandleBan,
	appmessage.CmdUnbanRequestMessage:                                       rpchandlers.HandleUnban,
	appmessage.CmdGetInfoRequestMessage:                                     rpchandlers.HandleGetInfo,
	appmessage.CmdNotifyPruningPointUTXOSetOverrideRequestMessage:           rpchandlers.HandleNotifyPruningPointUTXOSetOverrideRequest,
	appmessage.CmdStopNotifyingPruningPointUTXOSetOverrideRequestMessage:    rpchandlers.HandleStopNotifyingPruningPointUTXOSetOverrideRequest,
	appmessage.CmdEstimateNetworkHashesPerSecondRequestMessage:              rpchandlers.HandleEstimateNetworkHashesPerSecond,
	appmessage.CmdNotifyVirtualDaaScoreChangedRequestMessage:                rpchandlers.HandleNotifyVirtualDaaScoreChanged,
	appmessage.CmdNotifyNewBlockTemplateRequestMessage:                      rpchandlers.HandleNotifyNewBlockTemplate,
	appmessage.CmdGetMempoolEntriesByAddressesRequestMessage:                rpchandlers.HandleGetMempoolEntriesByAddresses,
}

func (m *Manager) routerInitializer(router *router.Router, netConnection *netadapter.NetConnection) {
	messageTypes := make([]appmessage.MessageCommand, 0, len(handlers))
	for messageType := range handlers {
		messageTypes = append(messageTypes, messageType)
	}
	incomingRoute, err := router.AddIncomingRoute("rpc router", messageTypes)
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

func (m *Manager) handleError(err error, netConnection *netadapter.NetConnection) {
	if errors.Is(err, router.ErrTimeout) {
		log.Warnf("Got timeout from %s. Disconnecting...", netConnection)
		netConnection.Disconnect()
		return
	}
	if errors.Is(err, router.ErrRouteClosed) {
		return
	}

	//INFO: `panic(err)` <- v. 0.12.0 kaspad Code that was here before, causing kaspad to crash with kaspactl.

	//TO DO: find out what to do with the err, BUT DO NOT PANIC due to an incomming message!!!

	//#######################################################################################
	//# 											#
	//#	Idea: 	1) find the appmessage response of the corrosponding request,		#
	//#		2) fill the rpc error field with the err,				#
	//#		3) process and send back to client.					#
	//#											#
	//#######################################################################################

	//for now just log - better then crashing kaspad, or doing nothing-
	log.Warnf("Got bad incoming message from %s. ", netConnection)

	//INFO: 1) trying to disconnect here causes tests to fail
	//	2) on kaspactl client side this casues `timeout of 30s has been exceeded` error after 30 secs.

}
