package rpcclient

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

type rpcRouter struct {
	router *routerpkg.Router
	routes map[appmessage.MessageCommand]*routerpkg.Route
}

func buildRPCRouter() (*rpcRouter, error) {
	router := routerpkg.NewRouter()
	routes := make(map[appmessage.MessageCommand]*routerpkg.Route, len(appmessage.RPCMessageCommandToString))
	for messageType := range appmessage.RPCMessageCommandToString {
		route, err := router.AddIncomingRoute("rpc", []appmessage.MessageCommand{messageType})
		if err != nil {
			return nil, err
		}
		routes[messageType] = route
	}

	return &rpcRouter{
		router: router,
		routes: routes,
	}, nil
}

func (r *rpcRouter) outgoingRoute() *routerpkg.Route {
	return r.router.OutgoingRoute()
}
