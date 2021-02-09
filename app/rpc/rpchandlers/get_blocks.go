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

	if !getBlocksRequest.IncludeBlockVerboseData && getBlocksRequest.IncludeTransactionVerboseData {
		return &appmessage.GetBlocksResponseMessage{
			Error: appmessage.RPCErrorf(
				"If includeTransactionVerboseData is set, then includeBlockVerboseData must be set as well"),
		}, nil
	}

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

	if len(blockHashes) < maxBlocksInGetBlocksResponse {
		virtualSelectedParentAnticone, err := context.Domain.Consensus().Anticone(virtualSelectedParent)
		if err != nil {
			return nil, err
		}
		nextLowHash = virtualSelectedParent
		blockHashes = append(blockHashes, virtualSelectedParentAnticone...)
	}

	if len(blockHashes) > maxBlocksInGetBlocksResponse {
		blockHashes = blockHashes[:maxBlocksInGetBlocksResponse]
	}

	response := &appmessage.GetBlocksResponseMessage{
		BlockHashes:      hashes.ToStrings(blockHashes),
		BlockVerboseData: make([]*appmessage.BlockVerboseData, len(blockHashes)),
		NextLowHash:      nextLowHash.String(),
	}

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
	return response, nil
}
