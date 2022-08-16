package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleStopNotifyingUTXOsChanged handles the respectively named RPC command
func HandleStopNotifyingUTXOsChanged(context *rpccontext.Context, router *router.Router, request appmessage.Message) (appmessage.Message, error) {

	stopNotifyingUTXOsChangedRequest := request.(*appmessage.StopNotifyingUTXOsChangedRequestMessage)

	if !context.Config.UTXOIndex {
		errorMessage := appmessage.NewStopNotifyingUTXOsChangedResponseMessage(stopNotifyingUTXOsChangedRequest.ID)
		errorMessage.Error = appmessage.RPCErrorf("Method unavailable when kaspad is run without --utxoindex")
		return errorMessage, nil
	}

	addresses, err := context.ConvertAddressStringsToUTXOsChangedNotificationAddresses(stopNotifyingUTXOsChangedRequest.Addresses)
	if err != nil {
		errorMessage := appmessage.NewNotifyUTXOsChangedResponseMessage(stopNotifyingUTXOsChangedRequest.ID)
		errorMessage.Error = appmessage.RPCErrorf("Parsing error: %s", err)
		return errorMessage, nil
	}

	listener, err := context.NotificationManager.Listener(router)
	if err != nil {
		return nil, err
	}
	listener.StopPropagatingUTXOsChangedNotifications(addresses)

	response := appmessage.NewStopNotifyingUTXOsChangedResponseMessage(stopNotifyingUTXOsChangedRequest.ID)
	return response, nil
}
