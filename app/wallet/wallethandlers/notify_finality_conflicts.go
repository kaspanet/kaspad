package wallethandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/wallet/walletcontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleNotifyFinalityConflicts handles the respectively named RPC command
func HandleNotifyFinalityConflicts(context walletcontext.Context, router *router.Router, _ appmessage.Message) (appmessage.Message, error) {
	listener, err := context.Listener(router)
	if err != nil {
		return nil, err
	}
	listener.PropagateFinalityConflictNotifications()
	listener.PropagateFinalityConflictResolvedNotifications()

	response := appmessage.NewNotifyFinalityConflictsResponseMessage()
	return response, nil
}
