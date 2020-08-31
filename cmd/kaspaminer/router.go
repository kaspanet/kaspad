package main

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server/grpcserver/grpcclient"
)

type minerRouter struct {
	router                        *routerpkg.Router
	getBlockTemplateResponseRoute *routerpkg.Route
	blockAddedNotificationRoute   *routerpkg.Route
}

func newRouter(rpcClient *grpcclient.RPCClient) (*minerRouter, error) {
	router := routerpkg.NewRouter()
	getBlockTemplateResponseRoute, err := router.AddIncomingRoute([]appmessage.MessageCommand{appmessage.CmdGetBlockTemplateResponseMessage})
	if err != nil {
		return nil, err
	}
	blockAddedNotificationRoute, err := router.AddIncomingRoute([]appmessage.MessageCommand{appmessage.CmdBlockAddedNotificationMessage})
	if err != nil {
		return nil, err
	}
	_, err = router.AddIncomingRoute([]appmessage.MessageCommand{appmessage.CmdNotifyBlockAddedResponseMessage, appmessage.CmdSubmitBlockResponseMessage})
	if err != nil {
		return nil, err
	}

	minerRouter := &minerRouter{
		router: router,

		getBlockTemplateResponseRoute: getBlockTemplateResponseRoute,
		blockAddedNotificationRoute:   blockAddedNotificationRoute,
	}

	spawn("NewRouter-sendLoop", func() {
		for {
			message, err := minerRouter.router.OutgoingRoute().Dequeue()
			if err != nil {
				panic(err)
			}
			err = rpcClient.SendAppMessage(message)
			if err != nil {
				panic(err)
			}
		}
	})
	spawn("NewRouter-receiveLoop", func() {
		for {
			message, err := rpcClient.ReceiveAppMessage()
			if err != nil {
				panic(err)
			}
			err = minerRouter.router.EnqueueIncomingMessage(message)
			if err != nil {
				panic(err)
			}
		}
	})

	return minerRouter, nil
}

func (r *minerRouter) outgoingRoute() *routerpkg.Route {
	return r.router.OutgoingRoute()
}
