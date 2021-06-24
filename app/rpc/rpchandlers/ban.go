package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"net"
)

// HandleBan handles the respectively named RPC command
func HandleBan(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	banRequest := request.(*appmessage.BanRequestMessage)
	ip := net.ParseIP(banRequest.IP)
	if ip == nil {
		hint := ""
		if banRequest.IP[0] == '[' {
			hint = " (try to remove “[” and “]” symbols)"
		}
		errorMessage := &appmessage.BanResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Could not parse IP%s: %s", hint, banRequest.IP)
		return errorMessage, nil
	}

	err := context.ConnectionManager.BanByIP(ip)
	if err != nil {
		errorMessage := &appmessage.BanResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Could not ban IP: %s", err)
		return errorMessage, nil
	}
	response := appmessage.NewBanResponseMessage()
	return response, nil
}
