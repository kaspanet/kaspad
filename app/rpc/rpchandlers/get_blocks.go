package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

const (
	// maxBlocksInGetBlocksResponse is the max amount of blocks that are
	// allowed in a GetBlocksResult.
	maxBlocksInGetBlocksResponse = 100
)

// HandleGetBlocks handles the respectively named RPC command
func HandleGetBlocks(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	response := &appmessage.GetBlocksResponseMessage{}
	response.Error = appmessage.RPCErrorf("not implemented")
	return response, nil
}
