package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/pkg/errors"
)

const (
	// maxBlocksInGetChainFromBlockResponse is the max amount of blocks that
	// are allowed in a GetChainFromBlockResponse.
	maxBlocksInGetChainFromBlockResponse = 1000
)

// HandleGetChainFromBlock handles the respectively named RPC command
func HandleGetChainFromBlock(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	getChainFromBlockRequest := request.(*appmessage.GetChainFromBlockRequestMessage)

	if context.AcceptanceIndex == nil {
		errorMessage := &appmessage.GetChainFromBlockResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("The acceptance index must be " +
			"enabled to get the selected parent chain " +
			"(specify --acceptanceindex)")
		return errorMessage, nil
	}

	var startHash *daghash.Hash
	if getChainFromBlockRequest.StartHash != "" {
		startHash = &daghash.Hash{}
		err := daghash.Decode(startHash, getChainFromBlockRequest.StartHash)
		if err != nil {
			errorMessage := &appmessage.GetChainFromBlockResponseMessage{}
			errorMessage.Error = appmessage.RPCErrorf("Could not parse startHash: %s", err)
			return errorMessage, nil
		}
	}

	context.DAG.RLock()
	defer context.DAG.RUnlock()

	// If startHash is not in the selected parent chain, there's nothing
	// to do; return an error.
	if startHash != nil && !context.DAG.IsInDAG(startHash) {
		errorMessage := &appmessage.GetChainFromBlockResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Block %s not found in the DAG", startHash)
		return errorMessage, nil
	}

	// Retrieve the selected parent chain.
	removedChainHashes, addedChainHashes, err := context.DAG.SelectedParentChain(startHash)
	if err != nil {
		return nil, err
	}

	// Limit the amount of blocks in the response
	if len(addedChainHashes) > maxBlocksInGetChainFromBlockResponse {
		addedChainHashes = addedChainHashes[:maxBlocksInGetChainFromBlockResponse]
	}

	// Collect addedChainBlocks.
	addedChainBlocks, err := context.CollectChainBlocks(addedChainHashes)
	if err != nil {
		errorMessage := &appmessage.GetChainFromBlockResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Could not collect chain blocks: %s", err)
		return errorMessage, nil
	}

	// Collect removedHashes.
	removedHashes := make([]string, len(removedChainHashes))
	for i, hash := range removedChainHashes {
		removedHashes[i] = hash.String()
	}

	// If the user specified to include the blocks, collect them as well.
	var blockVerboseData []*appmessage.BlockVerboseData
	if getChainFromBlockRequest.IncludeBlockVerboseData {
		data, err := hashesToBlockVerboseData(context, addedChainHashes)
		if err != nil {
			return nil, err
		}
		blockVerboseData = data
	}

	response := appmessage.NewGetChainFromBlockResponseMessage(removedHashes, addedChainBlocks, blockVerboseData)
	return response, nil
}

// hashesToBlockVerboseData takes block hashes and returns their
// correspondent block verbose.
func hashesToBlockVerboseData(context *rpccontext.Context, hashes []*daghash.Hash) ([]*appmessage.BlockVerboseData, error) {
	getBlockVerboseResults := make([]*appmessage.BlockVerboseData, 0, len(hashes))
	for _, blockHash := range hashes {
		block, err := context.DAG.BlockByHash(blockHash)
		if err != nil {
			return nil, errors.Errorf("could not retrieve block %s.", blockHash)
		}
		getBlockVerboseResult, err := context.BuildBlockVerboseData(block, false)
		if err != nil {
			return nil, errors.Wrapf(err, "could not build getBlockVerboseResult for block %s", blockHash)
		}
		getBlockVerboseResults = append(getBlockVerboseResults, getBlockVerboseResult)
	}
	return getBlockVerboseResults, nil
}
