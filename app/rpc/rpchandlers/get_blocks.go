package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

const (
	// maxBlocksInGetBlocksResponse is the max amount of blocks that are
	// allowed in a GetBlocksResult.
	maxBlocksInGetBlocksResponse = 100
)

// HandleGetBlocks handles the respectively named RPC command
func HandleGetBlocks(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	getBlocksRequest := request.(*appmessage.GetBlocksRequestMessage)

	// Validate that user didn't set IncludeTransactionVerboseData without setting IncludeBlockVerboseData
	if !getBlocksRequest.IncludeBlockVerboseData && getBlocksRequest.IncludeTransactionVerboseData {
		return &appmessage.GetBlocksResponseMessage{
			Error: appmessage.RPCErrorf(
				"If includeTransactionVerboseData is set, then includeBlockVerboseData must be set as well"),
		}, nil
	}

	// Decode lowHash
	// If lowHash is empty - use genesis instead.
	lowHash := context.Config.ActiveNetParams.GenesisHash
	if getBlocksRequest.LowHash != "" {
		var err error
		lowHash, err = externalapi.NewDomainHashFromString(getBlocksRequest.LowHash)
		if err != nil {
			return &appmessage.GetBlocksResponseMessage{
				Error: appmessage.RPCErrorf("Could not decode lowHash %s: %s", getBlocksRequest.LowHash, err),
			}, nil
		}
	}

	// Get hashes between lowHash and virtualSelectedParent
	virtualSelectedParent, err := context.Domain.Consensus().GetVirtualSelectedParent()
	if err != nil {
		return nil, err
	}
	blockHashes, err := context.Domain.Consensus().GetHashesBetween(
		lowHash, virtualSelectedParent, maxBlocksInGetBlocksResponse)
	if err != nil {
		return nil, err
	}
	nextLowHash := blockHashes[len(blockHashes)-1]

	// If there are no maxBlocksInGetBlocksResponse between lowHash and virtualSelectedParent -
	// add virtualSelectedParent's anticone
	if len(blockHashes) < maxBlocksInGetBlocksResponse {
		virtualSelectedParentAnticone, err := context.Domain.Consensus().Anticone(virtualSelectedParent)
		if err != nil {
			return nil, err
		}
		blockHashes = append(blockHashes, virtualSelectedParentAnticone...)

		// Don't move nextLowHash, since lowHash has to be in virtualSelectedParent's past
	}

	// Both GetHashesBetween and Anticone might return more then the allowed number of blocks, so
	// trim any extra blocks.
	if len(blockHashes) > maxBlocksInGetBlocksResponse {
		blockHashes = blockHashes[:maxBlocksInGetBlocksResponse]
	}

	// Prepare the response
	response := &appmessage.GetBlocksResponseMessage{
		BlockHashes: hashes.ToStrings(blockHashes),
		NextLowHash: nextLowHash.String(),
	}

	// Retrieve all block data in case BlockVerboseData was requested
	if getBlocksRequest.IncludeBlockVerboseData {
		response.BlockVerboseData = make([]*appmessage.BlockVerboseData, len(blockHashes))
		for i, blockHash := range blockHashes {
			blockHeader, err := context.Domain.Consensus().GetBlockHeader(blockHash)
			if err != nil {
				return nil, err
			}
			blockVerboseData, err := context.BuildBlockVerboseData(blockHeader, nil,
				getBlocksRequest.IncludeTransactionVerboseData)
			if err != nil {
				return nil, err
			}

			response.BlockVerboseData[i] = blockVerboseData
		}
	}
	return response, nil
}
