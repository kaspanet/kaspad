package wallethandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/wallet/walletcontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleNotifyBlockAdded handles the respectively named RPC command
func HandleNotifyBlockAdded(context walletcontext.Context, router *router.Router, _ appmessage.Message) (appmessage.Message, error) {
	listener, err := context.Listener(router)
	if err != nil {
		return nil, err
	}
	listener.PropagateBlockAddedNotifications()

	response := appmessage.NewNotifyBlockAddedResponseMessage()
	return response, nil
}
