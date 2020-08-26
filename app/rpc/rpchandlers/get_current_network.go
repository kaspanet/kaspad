package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleGetCurrentNetwork handles the respectively named RPC command
func HandleGetCurrentNetwork(context *rpccontext.Context, outgoingRoute *router.Route) error {
	log.Warnf("GOT CURRENT NET REQUEST")
	log.Warnf("HERE'S THE CURRENT NET: %s", context.DAG.Params.Name)

	message := appmessage.NewGetCurrentVersionResponseMessage(context.DAG.Params.Name)
	return outgoingRoute.Enqueue(message)
}
