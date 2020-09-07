package rpchandlers

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
)

const (
	// maxBlocksInGetBlocksResponse is the max amount of blocks that are
	// allowed in a GetBlocksResult.
	maxBlocksInGetBlocksResponse = 1000
)

// HandleGetBlocks handles the respectively named RPC command
func HandleGetBlocks(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	getBlocksRequest := request.(*appmessage.GetBlocksRequestMessage)

	var lowHash *daghash.Hash
	if getBlocksRequest.LowHash != "" {
		lowHash = &daghash.Hash{}
		err := daghash.Decode(lowHash, getBlocksRequest.LowHash)
		if err != nil {
			errorMessage := &appmessage.GetBlocksResponseMessage{}
			errorMessage.Error = appmessage.RPCErrorf("Could not parse lowHash: %s", err)
			return errorMessage, nil
		}
	}

	context.DAG.RLock()
	defer context.DAG.RUnlock()

	// If lowHash is not in the DAG, there's nothing to do; return an error.
	if lowHash != nil && !context.DAG.IsKnownBlock(lowHash) {
		errorMessage := &appmessage.GetBlocksResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Block %s not found in DAG", lowHash)
		return errorMessage, nil
	}

	// Retrieve the block hashes.
	blockHashes, err := context.DAG.BlockHashesFrom(lowHash, maxBlocksInGetBlocksResponse)
	if err != nil {
		return nil, err
	}

	// Convert the hashes to strings
	hashes := make([]string, len(blockHashes))
	for i, blockHash := range blockHashes {
		hashes[i] = blockHash.String()
	}

	// Include more data if requested
	var blockHexes []string
	var blockVerboseData []*appmessage.BlockVerboseData
	if getBlocksRequest.IncludeBlockHexes || getBlocksRequest.IncludeBlockVerboseData {
		blockBytesSlice, err := hashesToBlockBytes(context, blockHashes)
		if err != nil {
			return nil, err
		}
		if getBlocksRequest.IncludeBlockHexes {
			blockHexes = blockBytesToStrings(blockBytesSlice)
		}
		if getBlocksRequest.IncludeBlockVerboseData {
			data, err := blockBytesToBlockVerboseResults(context, blockBytesSlice, getBlocksRequest.IncludeBlockVerboseData)
			if err != nil {
				return nil, err
			}
			blockVerboseData = data
		}
	}

	response := appmessage.NewGetBlocksResponseMessage(hashes, blockHexes, blockVerboseData)
	return response, nil
}

func hashesToBlockBytes(context *rpccontext.Context, hashes []*daghash.Hash) ([][]byte, error) {
	blocks := make([][]byte, len(hashes))
	for i, hash := range hashes {
		block, err := context.DAG.BlockByHash(hash)
		if err != nil {
			return nil, err
		}
		blockBytes, err := block.Bytes()
		if err != nil {
			return nil, err
		}
		blocks[i] = blockBytes
	}
	return blocks, nil
}
func blockBytesToStrings(blockBytesSlice [][]byte) []string {
	rawBlocks := make([]string, len(blockBytesSlice))
	for i, blockBytes := range blockBytesSlice {
		rawBlocks[i] = hex.EncodeToString(blockBytes)
	}
	return rawBlocks
}

func blockBytesToBlockVerboseResults(context *rpccontext.Context, blockBytesSlice [][]byte,
	includeTransactionVerboseData bool) ([]*appmessage.BlockVerboseData, error) {

	verboseBlocks := make([]*appmessage.BlockVerboseData, len(blockBytesSlice))
	for i, blockBytes := range blockBytesSlice {
		block, err := util.NewBlockFromBytes(blockBytes)
		if err != nil {
			return nil, err
		}
		getBlockVerboseResult, err := context.BuildBlockVerboseData(block, includeTransactionVerboseData)
		if err != nil {
			return nil, err
		}
		verboseBlocks[i] = getBlockVerboseResult
	}
	return verboseBlocks, nil
}
