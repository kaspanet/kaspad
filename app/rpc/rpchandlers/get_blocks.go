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
	maxBlocksInGetBlocksResponse = 1000
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

		blockInfo, err := context.Domain.Consensus().GetBlockInfo(lowHash)
		if err != nil {
			return nil, err
		}

		if !blockInfo.Exists {
			return &appmessage.GetBlocksResponseMessage{
				Error: appmessage.RPCErrorf("Could not find lowHash %s", getBlocksRequest.LowHash),
			}, nil
		}
	}

	// Get hashes between lowHash and virtualSelectedParent
	virtualSelectedParent, err := context.Domain.Consensus().GetVirtualSelectedParent()
	if err != nil {
		return nil, err
	}
	blockHashes, highHash, err := context.Domain.Consensus().GetHashesBetween(lowHash, virtualSelectedParent, maxBlocksInGetBlocksResponse)
	if err != nil {
		return nil, err
	}

	// prepend low hash to make it inclusive
	blockHashes = append([]*externalapi.DomainHash{lowHash}, blockHashes...)

	// If the high hash is equal to virtualSelectedParent it means GetHashesBetween didn't skip any hashes, and
	// there's space to add the virtualSelectedParent's anticone, otherwise you can't add the anticone because
	// there's no guarantee that all of the anticone root ancestors will be present.
	if highHash.Equal(virtualSelectedParent) {
		virtualSelectedParentAnticone, err := context.Domain.Consensus().Anticone(virtualSelectedParent)
		if err != nil {
			return nil, err
		}
		blockHashes = append(blockHashes, virtualSelectedParentAnticone...)
	}

	// Both GetHashesBetween and Anticone might return more then the allowed number of blocks, so
	// trim any extra blocks.
	if len(blockHashes) > maxBlocksInGetBlocksResponse {
		blockHashes = blockHashes[:maxBlocksInGetBlocksResponse]
	}

	// Prepare the response
	response := &appmessage.GetBlocksResponseMessage{
		BlockHashes: hashes.ToStrings(blockHashes),
	}

	// Retrieve all block data in case RPCBlockVerboseData was requested
	if getBlocksRequest.IncludeBlockVerboseData {
		response.BlockVerboseData = make([]*appmessage.RPCBlockVerboseData, len(blockHashes))
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
