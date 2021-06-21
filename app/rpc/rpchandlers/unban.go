package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"net"
)

// HandleUnban handles the respectively named RPC command
func HandleUnban(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	unbanRequest := request.(*appmessage.UnbanRequestMessage)
	ip := net.ParseIP(unbanRequest.IP)
	if ip == nil {
		hint := ""
		if unbanRequest.IP[0] == '[' {
			hint = " (try to remove “[” and “]” symbols)"
		}
		errorMessage := &appmessage.UnbanResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Could not parse IP%s: %s", hint, unbanRequest.IP)
		return errorMessage, nil
	}
	err := context.AddressManager.Unban(appmessage.NewNetAddressIPPort(ip, 0))
	if err != nil {
		errorMessage := &appmessage.UnbanResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Could not unban IP: %s", err)
		return errorMessage, nil
	}
	response := appmessage.NewUnbanResponseMessage()
	return response, nil
}
