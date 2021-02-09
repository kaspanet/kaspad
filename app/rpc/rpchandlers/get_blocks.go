package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/model"
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
	blockHashes, err := context.Domain.Consensus().GetHashesBetween(lowHash, model.VirtualBlockHash,
		maxBlocksInGetBlocksResponse)
	if err != nil {
		return nil, err
	}

	if len(blockHashes) > maxBlocksInGetBlocksResponse {
		blockHashes = blockHashes[:maxBlocksInGetBlocksResponse]
	}

	response := &appmessage.GetBlocksResponseMessage{
		BlockHashes:      hashes.ToStrings(blockHashes),
		BlockVerboseData: make([]*appmessage.BlockVerboseData, len(blockHashes)),
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
