package rpchandlers

import (
	"github.com/zoomy-network/zoomyd/app/appmessage"
	"github.com/zoomy-network/zoomyd/app/rpc/rpccontext"
	"github.com/zoomy-network/zoomyd/infrastructure/network/netadapter/router"
)

// HandleNotifyUTXOsChanged handles the respectively named RPC command
func HandleNotifyUTXOsChanged(context *rpccontext.Context, router *router.Router, request appmessage.Message) (appmessage.Message, error) {
	if !context.Config.UTXOIndex {
		errorMessage := appmessage.NewNotifyUTXOsChangedResponseMessage()
		errorMessage.Error = appmessage.RPCErrorf("Method unavailable when zoomyd is run without --utxoindex")
		return errorMessage, nil
	}

	notifyUTXOsChangedRequest := request.(*appmessage.NotifyUTXOsChangedRequestMessage)
	addresses, err := context.ConvertAddressStringsToUTXOsChangedNotificationAddresses(notifyUTXOsChangedRequest.Addresses)
	if err != nil {
		errorMessage := appmessage.NewNotifyUTXOsChangedResponseMessage()
		errorMessage.Error = appmessage.RPCErrorf("Parsing error: %s", err)
		return errorMessage, nil
	}

	listener, err := context.NotificationManager.Listener(router)
	if err != nil {
		return nil, err
	}
	context.NotificationManager.PropagateUTXOsChangedNotifications(listener, addresses)

	response := appmessage.NewNotifyUTXOsChangedResponseMessage()
	return response, nil
}
