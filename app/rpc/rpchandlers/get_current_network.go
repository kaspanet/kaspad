package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleGetCurrentNetwork handles the respectively named RPC command
func HandleGetCurrentNetwork(context *rpccontext.Context, outgoingRoute *router.Route) error {
	message := appmessage.NewGetCurrentNetworkResponseMessage(context.DAG.Params.Net.String())
	return outgoingRoute.Enqueue(message)
}
