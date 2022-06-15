package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleNotifyNewBlockTemplate handles the respectively named RPC command
func HandleNotifyNewBlockTemplate(context *rpccontext.Context, router *router.Router, request appmessage.Message) (appmessage.Message, error) {
	
	notifyNewBlockTemplateRequest := request.(*appmessage.NotifyNewBlockTemplateRequestMessage)
	
	listener, err := context.NotificationManager.Listener(router)
	if err != nil {
		return nil, err
	}
	listener.PropagateNewBlockTemplateNotifications(notifyNewBlockTemplateRequest.Id)

	response := appmessage.NewNotifyNewBlockTemplateResponseMessage(notifyNewBlockTemplateRequest.Id)
	return response, nil
}
