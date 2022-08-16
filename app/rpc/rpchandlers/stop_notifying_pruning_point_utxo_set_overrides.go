package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleStopNotifyingPruningPointUTXOSetOverrideRequest handles the respectively named RPC command
func HandleStopNotifyingPruningPointUTXOSetOverrideRequest(context *rpccontext.Context, router *router.Router, request appmessage.Message) (appmessage.Message, error) {

	stopNotifyingPruningPointUTXOSetOverrideRequest := request.(*appmessage.StopNotifyingPruningPointUTXOSetOverrideRequestMessage)

	listener, err := context.NotificationManager.Listener(router)
	if err != nil {
		return nil, err
	}
	listener.StopPropagatingPruningPointUTXOSetOverrideNotifications(stopNotifyingPruningPointUTXOSetOverrideRequest.ID)

	response := appmessage.NewStopNotifyingPruningPointUTXOSetOverrideResponseMessage(stopNotifyingPruningPointUTXOSetOverrideRequest.ID)
	return response, nil
}
