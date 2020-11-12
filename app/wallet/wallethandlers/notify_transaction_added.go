package wallethandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/wallet/walletcontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleNotifyTransactionAdded handles the respectively named RPC command
func HandleNotifyTransactionAdded(context walletcontext.Context, router *router.Router, request appmessage.Message) (appmessage.Message, error) {
	listener, err := context.Listener(router)
	if err != nil {
		return nil, err
	}

	transactionAddedRequestMessage := request.(*appmessage.NotifyTransactionAddedRequestMessage)
	listener.PropagateTransactionAddedNotifications(transactionAddedRequestMessage.Transaction.TxHash())

	response := appmessage.NewNotifyTransactionAddedResponseMessage()
	return response, nil
}
