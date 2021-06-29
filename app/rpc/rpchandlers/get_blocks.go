package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleGetBlocks handles the respectively named RPC command
func HandleGetBlocks(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	getBlocksRequest := request.(*appmessage.GetBlocksRequestMessage)

	// Validate that user didn't set IncludeTransactions without setting IncludeBlocks
	if !getBlocksRequest.IncludeBlocks && getBlocksRequest.IncludeTransactions {
		return &appmessage.GetBlocksResponseMessage{
			Error: appmessage.RPCErrorf(
				"If includeTransactions is set, then includeBlockVerboseData must be set as well"),
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

	// We use +1 because lowHash also returns
	maxBlocks := context.Config.NetParams().MergeSetSizeLimit + 1
	blockHashes, highHash, err := context.Domain.Consensus().GetHashesBetween(lowHash, virtualSelectedParent, maxBlocks)
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

	// Prepare the response
	response := appmessage.NewGetBlocksResponseMessage()
	response.BlockHashes = hashes.ToStrings(blockHashes)
	if getBlocksRequest.IncludeBlocks {
		rpcBlocks := make([]*appmessage.RPCBlock, len(blockHashes))
		for i, blockHash := range blockHashes {
			block, err := context.Domain.Consensus().GetBlockEvenIfHeaderOnly(blockHash)
			if err != nil {
				return nil, err
			}

			if getBlocksRequest.IncludeTransactions {
				rpcBlocks[i] = appmessage.DomainBlockToRPCBlock(block)
			} else {
				rpcBlocks[i] = appmessage.DomainBlockToRPCBlock(&externalapi.DomainBlock{Header: block.Header})
			}
			err = context.PopulateBlockWithVerboseData(rpcBlocks[i], block.Header, nil, getBlocksRequest.IncludeTransactions)
			if err != nil {
				return nil, err
			}
		}
		response.Blocks = rpcBlocks
	}

	return response, nil
}
