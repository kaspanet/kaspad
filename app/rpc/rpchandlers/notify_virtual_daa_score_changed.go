package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleNotifyVirtualDaaScoreChanged handles the respectively named RPC command
func HandleNotifyVirtualDaaScoreChanged(context *rpccontext.Context, router *router.Router, request appmessage.Message) (appmessage.Message, error) {

	notifyVirtualDaaScoreChangedRequest := request.(*appmessage.NotifyVirtualDaaScoreChangedRequestMessage)

	listener, err := context.NotificationManager.Listener(router)
	if err != nil {
		return nil, err
	}
	listener.PropagateVirtualDaaScoreChangedNotifications(notifyVirtualDaaScoreChangedRequest.ID)

	response := appmessage.NewNotifyVirtualDaaScoreChangedResponseMessage(notifyVirtualDaaScoreChangedRequest.ID)
	return response, nil
}
