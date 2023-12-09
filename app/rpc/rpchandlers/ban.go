package rpchandlers

import (
<<<<<<< Updated upstream
	"github.com/zoomy-network/zoomyd/app/appmessage"
	"github.com/zoomy-network/zoomyd/app/rpc/rpccontext"
	"github.com/zoomy-network/zoomyd/infrastructure/network/netadapter/router"
=======
>>>>>>> Stashed changes
	"net"

	"github.com/zoomy-network/zoomyd/app/appmessage"
	"github.com/zoomy-network/zoomyd/app/rpc/rpccontext"
	"github.com/zoomy-network/zoomyd/infrastructure/network/netadapter/router"
)

// HandleBan handles the respectively named RPC command
func HandleBan(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	if context.Config.SafeRPC {
		log.Warn("Ban RPC command called while node in safe RPC mode -- ignoring.")
		response := appmessage.NewBanResponseMessage()
		response.Error =
			appmessage.RPCErrorf("Ban RPC command called while node in safe RPC mode")
		return response, nil
	}

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
