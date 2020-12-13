package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util"
)

// HandleNotifyChainChanged handles the respectively named RPC command
func HandleNotifyUTXOsChanged(context *rpccontext.Context, router *router.Router, request appmessage.Message) (appmessage.Message, error) {
	notifyUTXOsChangedRequest := request.(*appmessage.NotifyUTXOsChangedRequestMessage)

	scriptPublicKeys := make([][]byte, len(notifyUTXOsChangedRequest.Addresses))
	for i, addressString := range notifyUTXOsChangedRequest.Addresses {
		address, err := util.DecodeAddress(addressString, context.Config.ActiveNetParams.Prefix)
		if err != nil {
			errorMessage := appmessage.NewNotifyUTXOsChangedResponseMessage()
			errorMessage.Error = appmessage.RPCErrorf("Could not decode address '%s': %s", addressString, err)
			return errorMessage, nil
		}
		scriptPublicKey, err := txscript.PayToAddrScript(address)
		if err != nil {
			errorMessage := appmessage.NewNotifyUTXOsChangedResponseMessage()
			errorMessage.Error = appmessage.RPCErrorf("Could not create a scriptPublicKey for address '%s': %s", addressString, err)
			return errorMessage, nil
		}
		scriptPublicKeys[i] = scriptPublicKey
	}

	listener, err := context.NotificationManager.Listener(router)
	if err != nil {
		return nil, err
	}
	listener.PropagateUTXOsChangedNotifications(scriptPublicKeys)

	response := appmessage.NewNotifyUTXOsChangedResponseMessage()
	return response, nil
}
