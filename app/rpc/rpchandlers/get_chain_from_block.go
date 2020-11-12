package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

const (
	// maxBlocksInGetChainFromBlockResponse is the max amount of blocks that
	// are allowed in a GetChainFromBlockResponse.
	maxBlocksInGetChainFromBlockResponse = 1000
)

// HandleGetChainFromBlock handles the respectively named RPC command
func HandleGetChainFromBlock(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	return &appmessage.GetChainFromBlockResponseMessage{}, nil
}
