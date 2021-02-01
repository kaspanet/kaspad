package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// Unban handles the respectively named RPC command
func Unban(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	unbanRequest := request.(*appmessage.UnbanRequestMessage)
	err := context.AddressManager.Unban(appmessage.NewNetAddressIPPort(unbanRequest.IP, 0, 0))
	if err != nil {
		errorMessage := &appmessage.UnbanResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Could not unban IP: %s", err)
		return errorMessage, nil
	}
	response := appmessage.NewUnbanResponseMessage()
	return response, nil
}
