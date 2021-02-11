package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleStopNotifyingUTXOsChanged handles the respectively named RPC command
func HandleStopNotifyingUTXOsChanged(context *rpccontext.Context, router *router.Router, request appmessage.Message) (appmessage.Message, error) {
	if !context.Config.UTXOIndex {
		errorMessage := appmessage.NewStopNotifyingUTXOsChangedResponseMessage()
		errorMessage.Error = appmessage.RPCErrorf("Method unavailable when kaspad is run without --utxoindex")
		return errorMessage, nil
	}

	StopNotifyingUTXOsChangedRequest := request.(*appmessage.StopNotifyingUTXOsChangedRequestMessage)

	listener, err := context.NotificationManager.Listener(router)
	if err != nil {
		return nil, err
	}
	listener.StopPropagatingUTXOsChangedNotifications(StopNotifyingUTXOsChangedRequest.Addresses)

	response := appmessage.NewStopNotifyingUTXOsChangedResponseMessage()
	return response, nil
}
