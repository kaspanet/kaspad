package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleNotifyBlockAdded handles the respectively named RPC command
func HandleNotifyBlockAdded(context *rpccontext.Context, router *router.Router, request appmessage.Message) (appmessage.Message, error) {

	notifyBlockAddedRequestMessage := request.(*appmessage.NotifyBlockAddedRequestMessage)

	listener, err := context.NotificationManager.Listener(router)
	if err != nil {
		return nil, err
	}
	listener.PropagateBlockAddedNotifications(notifyBlockAddedRequestMessage.ID)

	response := appmessage.NewNotifyBlockAddedResponseMessage(notifyBlockAddedRequestMessage.ID)
	return response, nil
}
