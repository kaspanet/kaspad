package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleNotifyPruningPointUTXOSetOverrideRequest handles the respectively named RPC command
func HandleNotifyPruningPointUTXOSetOverrideRequest(context *rpccontext.Context, router *router.Router, request appmessage.Message) (appmessage.Message, error) {

	notifyPruningPointUTXOSetOverrideRequest := request.(*appmessage.NotifyPruningPointUTXOSetOverrideRequestMessage)

	listener, err := context.NotificationManager.Listener(router)
	if err != nil {
		return nil, err
	}
	listener.PropagatePruningPointUTXOSetOverrideNotifications(notifyPruningPointUTXOSetOverrideRequest.ID)

	response := appmessage.NewNotifyPruningPointUTXOSetOverrideResponseMessage(notifyPruningPointUTXOSetOverrideRequest.ID)
	return response, nil
}
