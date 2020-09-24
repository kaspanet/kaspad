package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util/daghash"
)

// HandleResolveFinalityConflict handles the respectively named RPC command
func HandleResolveFinalityConflict(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	ResolveFinalityConflictRequest := request.(*appmessage.ResolveFinalityConflictRequestMessage)

	finalityBlockHash, err := daghash.NewHashFromStr(ResolveFinalityConflictRequest.FinalityBlockHash)
	if err != nil {
		errorMessage := &appmessage.ResolveFinalityConflictResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Could not parse finalityBlockHash: %s", err)
		return errorMessage, nil
	}

	err = context.DAG.ResolveFinalityConflict(finalityBlockHash)
	if err != nil {
		errorMessage := &appmessage.ResolveFinalityConflictResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Could not resolve finality conflict: %s", err)
		return errorMessage, nil
	}

	response := appmessage.NewResolveFinalityConflictResponseMessage()
	return response, nil
}
