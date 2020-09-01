package main

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

type minerRouter struct {
	router                        *routerpkg.Router
	getBlockTemplateResponseRoute *routerpkg.Route
	submitBlockResponseRoute      *routerpkg.Route
	notifyBlockAddedResponseRoute *routerpkg.Route
	blockAddedNotificationRoute   *routerpkg.Route
}

func buildRouter() (*minerRouter, error) {
	router := routerpkg.NewRouter()
	getBlockTemplateResponseRoute, err := router.AddIncomingRoute([]appmessage.MessageCommand{appmessage.CmdGetBlockTemplateResponseMessage})
	if err != nil {
		return nil, err
	}
	submitBlockResponseRoute, err := router.AddIncomingRoute([]appmessage.MessageCommand{appmessage.CmdSubmitBlockResponseMessage})
	if err != nil {
		return nil, err
	}
	notifyBlockAddedResponseRoute, err := router.AddIncomingRoute([]appmessage.MessageCommand{appmessage.CmdNotifyBlockAddedResponseMessage})
	if err != nil {
		return nil, err
	}
	blockAddedNotificationRoute, err := router.AddIncomingRoute([]appmessage.MessageCommand{appmessage.CmdBlockAddedNotificationMessage})
	if err != nil {
		return nil, err
	}

	minerRouter := &minerRouter{
		router: router,

		getBlockTemplateResponseRoute: getBlockTemplateResponseRoute,
		submitBlockResponseRoute:      submitBlockResponseRoute,
		notifyBlockAddedResponseRoute: notifyBlockAddedResponseRoute,
		blockAddedNotificationRoute:   blockAddedNotificationRoute,
	}

	return minerRouter, nil
}

func (r *minerRouter) outgoingRoute() *routerpkg.Route {
	return r.router.OutgoingRoute()
}
