package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
)

// HandleGetBlockTemplate handles the respectively named RPC command
func HandleGetBlockTemplate(context *rpccontext.Context, request appmessage.Message) (appmessage.Message, error) {
	response := appmessage.NewGetBlockTemplateResponseMessage()
	return response, nil
}
