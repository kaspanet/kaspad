package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util"
)

// HandleNotifyBlockAdded handles the respectively named RPC command
func HandleNotifyBlockAdded(context *rpccontext.Context, router *router.Router, _ appmessage.Message) (appmessage.Message, error) {
	listener, err := context.NotificationManager.Listener(router)
	if err != nil {
		return nil, err
	}
	listener.SetOnBlockAddedListener(func(block *util.Block) error {
		notification := appmessage.NewBlockAddedNotificationMessage(block.MsgBlock())
		return router.OutgoingRoute().Enqueue(notification)
	})

	response := appmessage.NewNotifyBlockAddedResponseMessage()
	return response, nil
}
