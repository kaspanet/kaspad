package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleNotifyFinalityConflicts handles the respectively named RPC command
func HandleNotifyFinalityConflicts(context *rpccontext.Context, router *router.Router, request appmessage.Message) (appmessage.Message, error) {
	
	notifyFinalityConflictsRequest := request.(*appmessage.NotifyFinalityConflictsRequestMessage)
	
	listener, err := context.NotificationManager.Listener(router)
	if err != nil {
		return nil, err
	}
	listener.PropagateFinalityConflictNotifications(notifyFinalityConflictsRequest.Id)
	listener.PropagateFinalityConflictResolvedNotifications(notifyFinalityConflictsRequest.Id)

	response := appmessage.NewNotifyFinalityConflictsResponseMessage(notifyFinalityConflictsRequest.Id)
	return response, nil
}
