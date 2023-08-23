package rpchandlers

import (
	"github.com/c4ei/yunseokyeol/app/appmessage"
	"github.com/c4ei/yunseokyeol/app/rpc/rpccontext"
	"github.com/c4ei/yunseokyeol/infrastructure/network/netadapter/router"
)

// HandleNotifyFinalityConflicts handles the respectively named RPC command
func HandleNotifyFinalityConflicts(context *rpccontext.Context, router *router.Router, _ appmessage.Message) (appmessage.Message, error) {
	listener, err := context.NotificationManager.Listener(router)
	if err != nil {
		return nil, err
	}
	listener.PropagateFinalityConflictNotifications()
	listener.PropagateFinalityConflictResolvedNotifications()

	response := appmessage.NewNotifyFinalityConflictsResponseMessage()
	return response, nil
}
