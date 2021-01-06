package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util/network"
)

// HandleAddPeer handles the respectively named RPC command
func HandleAddPeer(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	AddPeerRequest := request.(*appmessage.AddPeerRequestMessage)
	address, err := network.NormalizeAddress(AddPeerRequest.Address, context.Config.ActiveNetParams.DefaultPort)
	if err != nil {
		errorMessage := &appmessage.AddPeerResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Could not parse address: %s", err)
		return errorMessage, nil
	}

	context.ConnectionManager.AddConnectionRequest(address, AddPeerRequest.IsPermanent)

	response := appmessage.NewAddPeerResponseMessage()
	return response, nil
}
