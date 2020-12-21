package rpchandlers

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleGetVirtualSelectedParentChainFromBlock handles the respectively named RPC command
func HandleGetVirtualSelectedParentChainFromBlock(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	getVirtualSelectedParentChainFromBlockRequest := request.(*appmessage.GetVirtualSelectedParentChainFromBlockRequestMessage)

	startHashBytes, err := hex.DecodeString(getVirtualSelectedParentChainFromBlockRequest.StartHash)
	if err != nil {
		errorMessage := &appmessage.GetVirtualSelectedParentChainFromBlockResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Could not parse startHash: %s", err)
		return errorMessage, nil
	}
	startHash, err := hashes.FromBytes(startHashBytes)
	if err != nil {
		errorMessage := &appmessage.GetVirtualSelectedParentChainFromBlockResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Could not parse startHash: %s", err)
		return errorMessage, nil
	}

	virtualSelectedParentChain, err := context.Domain.Consensus().GetVirtualSelectedParentChainFromBlock(startHash)
	if err != nil {
		response := &appmessage.GetVirtualSelectedParentChainFromBlockResponseMessage{}
		response.Error = appmessage.RPCErrorf("Could not build virtual "+
			"selected parent chain from %s: %s", getVirtualSelectedParentChainFromBlockRequest.StartHash, err)
		return response, nil
	}

	chainChangedNotification, err := context.ConvertVirtualSelectedParentChainChangesToChainChangedNotificationMessage(virtualSelectedParentChain)
	if err != nil {
		return nil, err
	}

	response := appmessage.NewGetVirtualSelectedParentChainFromBlockResponseMessage(
		chainChangedNotification.RemovedChainBlockHashes, chainChangedNotification.AddedChainBlocks)
	return response, nil
}
