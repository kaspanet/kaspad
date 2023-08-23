package rpchandlers

import (
	"github.com/c4ei/YunSeokYeol/app/appmessage"
	"github.com/c4ei/YunSeokYeol/app/rpc/rpccontext"
	"github.com/c4ei/YunSeokYeol/infrastructure/network/netadapter/router"
)

// HandleGetCurrentNetwork handles the respectively named RPC command
func HandleGetCurrentNetwork(context *rpccontext.Context, _ *router.Router, _ appmessage.Message) (appmessage.Message, error) {
	response := appmessage.NewGetCurrentNetworkResponseMessage(context.Config.ActiveNetParams.Net.String())
	return response, nil
}
