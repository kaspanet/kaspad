package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

const (
	// maxBlocksInGetVirtualSelectedParentChainFromBlockResponse is the max amount of blocks that
	// are allowed in a GetVirtualSelectedParentChainFromBlockResponse.
	maxBlocksInGetVirtualSelectedParentChainFromBlockResponse = 1000
)

// HandleGetVirtualSelectedParentChainFromBlock handles the respectively named RPC command
func HandleGetVirtualSelectedParentChainFromBlock(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	response := &appmessage.GetVirtualSelectedParentChainFromBlockResponseMessage{}
	response.Error = appmessage.RPCErrorf("not implemented")
	return response, nil
}
