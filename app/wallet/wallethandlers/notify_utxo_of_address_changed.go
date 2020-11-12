package wallethandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/wallet/walletcontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleNotifyUTXOOfAddressChanged handles the respectively named RPC command
func HandleNotifyUTXOOfAddressChanged(context walletcontext.Context, router *router.Router, request appmessage.Message) (appmessage.Message, error) {
	listener, err := context.Listener(router)
	if err != nil {
		return nil, err
	}

	requestMessage := request.(*appmessage.NotifyUTXOOfAddressChangedRequestMessage)
	listener.PropagateUTXOOfAddressChangedNotifications(requestMessage.Addresses)

	response := appmessage.NewNotifyUTXOOfAddressChangedResponseMessage()
	return response, nil
}
