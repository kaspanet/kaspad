package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleGetBlockDAGInfo handles the respectively named RPC command
func HandleGetBlockDAGInfo(context *rpccontext.Context, _ *router.Router, _ appmessage.Message) (appmessage.Message, error) {
	params := context.Config.ActiveNetParams

	response := appmessage.NewGetBlockDAGInfoResponseMessage()
	response.NetworkName = params.Name

	return response, nil
}
