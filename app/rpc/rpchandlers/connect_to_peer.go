package rpchandlers

import (
	"fmt"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util/network"
)

// HandleConnectToPeer handles the respectively named RPC command
func HandleConnectToPeer(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	connectToPeerRequest := request.(*appmessage.ConnectToPeerRequestMessage)
	address, err := network.NormalizeAddress(connectToPeerRequest.Address, context.DAG.Params.DefaultPort)
	if err != nil {
		errorMessage := &appmessage.ConnectToPeerResponseMessage{}
		errorMessage.Error = &appmessage.RPCError{
			Message: fmt.Sprintf("Could not parse address: %s", err),
		}
		return errorMessage, nil
	}

	context.ConnectionManager.AddConnectionRequest(address, connectToPeerRequest.IsPermanent)

	response := appmessage.NewConnectToPeerResponseMessage()
	return response, nil
}
