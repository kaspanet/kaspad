package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleNotifyVirtualSelectedParentBlueScoreChanged handles the respectively named RPC command
func HandleNotifyVirtualSelectedParentBlueScoreChanged(context *rpccontext.Context, router *router.Router, request appmessage.Message) (appmessage.Message, error) {
	
	notifyVirtualSelectedParentBlueScoreChangedRequest := request.(*appmessage.NotifyVirtualSelectedParentBlueScoreChangedRequestMessage)
	
	listener, err := context.NotificationManager.Listener(router)
	if err != nil {
		return nil, err
	}
	listener.PropagateVirtualSelectedParentBlueScoreChangedNotifications(notifyVirtualSelectedParentBlueScoreChangedRequest.Id)

	response := appmessage.NewNotifyVirtualSelectedParentBlueScoreChangedResponseMessage(notifyVirtualSelectedParentBlueScoreChangedRequest.Id)
	return response, nil
}
