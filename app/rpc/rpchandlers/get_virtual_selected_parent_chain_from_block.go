package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleGetVirtualSelectedParentChainFromBlock handles the respectively named RPC command
func HandleGetVirtualSelectedParentChainFromBlock(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	chainRequest := request.(*appmessage.GetVirtualSelectedParentChainFromBlockRequestMessage)

	startHash, err := externalapi.NewDomainHashFromString(chainRequest.StartHash)
	if err != nil {
		errorMessage := &appmessage.GetVirtualSelectedParentChainFromBlockResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Could not parse startHash: %s", err)
		return errorMessage, nil
	}

	virtualSelectedParentChain, err := context.Domain.Consensus().GetVirtualSelectedParentChainFromBlock(startHash)
	if err != nil {
		response := &appmessage.GetVirtualSelectedParentChainFromBlockResponseMessage{}
		response.Error = appmessage.RPCErrorf("Could not build virtual "+
			"selected parent chain from %s: %s", chainRequest.StartHash, err)
		return response, nil
	}

	if chainRequest.BatchSize > 0 && uint64(len(virtualSelectedParentChain.Added)) > chainRequest.BatchSize {
		// Send at most `BatchSize` added chain blocks
		virtualSelectedParentChain.Added = virtualSelectedParentChain.Added[:chainRequest.BatchSize]
	}

	chainChangedNotification, err := context.ConvertVirtualSelectedParentChainChangesToChainChangedNotificationMessage(
		virtualSelectedParentChain, chainRequest.IncludeAcceptedTransactionIDs)
	if err != nil {
		response := &appmessage.GetVirtualSelectedParentChainFromBlockResponseMessage{}
		response.Error = appmessage.RPCErrorf("Could not load acceptance data for virtual "+
			"selected parent chain from %s: %s", chainRequest.StartHash, err)
		return response, nil
	}

	response := appmessage.NewGetVirtualSelectedParentChainFromBlockResponseMessage(
		chainChangedNotification.RemovedChainBlockHashes, chainChangedNotification.AddedChainBlockHashes,
		chainChangedNotification.AcceptedTransactionIDs)
	return response, nil
}
