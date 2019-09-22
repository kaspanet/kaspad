package rpc

import (
	"fmt"
	"github.com/daglabs/btcd/btcjson"
	"github.com/daglabs/btcd/util/daghash"
)

// handleGetChainFromBlock implements the getChainFromBlock command.
func handleGetChainFromBlock(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	if s.cfg.AcceptanceIndex == nil {
		return nil, &btcjson.RPCError{
			Code: btcjson.ErrRPCNoAcceptanceIndex,
			Message: "The acceptance index must be " +
				"enabled to get the selected parent chain " +
				"(specify --acceptanceindex)",
		}
	}

	c := cmd.(*btcjson.GetChainFromBlockCmd)
	var startHash *daghash.Hash
	if c.StartHash != nil {
		startHash = &daghash.Hash{}
		err := daghash.Decode(startHash, *c.StartHash)
		if err != nil {
			return nil, rpcDecodeHexError(*c.StartHash)
		}
	}

	s.cfg.DAG.RLock()
	defer s.cfg.DAG.RUnlock()

	// If startHash is not in the selected parent chain, there's nothing
	// to do; return an error.
	if startHash != nil && !s.cfg.DAG.BlockExists(startHash) {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCBlockNotFound,
			Message: "Block not found in the DAG",
		}
	}

	// Retrieve the selected parent chain.
	removedChainHashes, addedChainHashes, err := s.cfg.DAG.SelectedParentChain(startHash)
	if err != nil {
		return nil, err
	}

	// Limit the amount of blocks in the response
	if len(addedChainHashes) > maxBlocksInGetChainFromBlockResult {
		addedChainHashes = addedChainHashes[:maxBlocksInGetChainFromBlockResult]
	}

	// Collect addedChainBlocks.
	addedChainBlocks, err := collectChainBlocks(s, addedChainHashes)
	if err != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCInternal.Code,
			Message: fmt.Sprintf("could not collect chain blocks: %s", err),
		}
	}

	// Collect removedHashes.
	removedHashes := make([]string, len(removedChainHashes))
	for i, hash := range removedChainHashes {
		removedHashes[i] = hash.String()
	}

	result := &btcjson.GetChainFromBlockResult{
		RemovedChainBlockHashes: removedHashes,
		AddedChainBlocks:        addedChainBlocks,
		Blocks:                  nil,
	}

	// If the user specified to include the blocks, collect them as well.
	if c.IncludeBlocks {
		getBlockVerboseResults, err := hashesToGetBlockVerboseResults(s, addedChainHashes)
		if err != nil {
			return nil, err
		}
		result.Blocks = getBlockVerboseResults
	}

	return result, nil
}
