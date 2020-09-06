package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleNotifyChainChanged handles the respectively named RPC command
func HandleNotifyChainChanged(context *rpccontext.Context, router *router.Router, _ appmessage.Message) (appmessage.Message, error) {
	if context.AcceptanceIndex == nil {
		errorMessage := appmessage.NewNotifyChainChangedResponseMessage()
		errorMessage.Error = &appmessage.RPCError{
			Message: "Acceptance index is not available",
		}
		return errorMessage, nil
	}

	listener, err := context.NotificationManager.Listener(router)
	if err != nil {
		return nil, err
	}
	listener.SetOnChainChangedListener(func(message *appmessage.ChainChangedNotificationMessage) error {
		return router.OutgoingRoute().Enqueue(message)
	})

	response := appmessage.NewNotifyChainChangedResponseMessage()
	return response, nil
}
