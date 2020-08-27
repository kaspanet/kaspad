package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleSubmitBlock handles the respectively named RPC command
func HandleSubmitBlock(context *rpccontext.Context, outgoingRoute *router.Route) error {
	message := appmessage.NewSubmitBlockResponseMessage()
	return outgoingRoute.Enqueue(message)
}
