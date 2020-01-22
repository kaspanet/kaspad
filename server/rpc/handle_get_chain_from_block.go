package rpc

import (
	"fmt"
	"github.com/kaspanet/kaspad/rpcmodel"
	"github.com/kaspanet/kaspad/util/daghash"
)

const (
	// maxBlocksInGetChainFromBlockResult is the max amount of blocks that
	// are allowed in a GetChainFromBlockResult.
	maxBlocksInGetChainFromBlockResult = 1000
)

// handleGetChainFromBlock implements the getChainFromBlock command.
func handleGetChainFromBlock(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	if s.cfg.AcceptanceIndex == nil {
		return nil, &rpcmodel.RPCError{
			Code: rpcmodel.ErrRPCNoAcceptanceIndex,
			Message: "The acceptance index must be " +
				"enabled to get the selected parent chain " +
				"(specify --acceptanceindex)",
		}
	}

	c := cmd.(*rpcmodel.GetChainFromBlockCmd)
	var lowHash *daghash.Hash
	if c.LowHash != nil {
		lowHash = &daghash.Hash{}
		err := daghash.Decode(lowHash, *c.LowHash)
		if err != nil {
			return nil, rpcDecodeHexError(*c.LowHash)
		}
	}

	s.cfg.DAG.RLock()
	defer s.cfg.DAG.RUnlock()

	// If lowHash is not in the selected parent chain, there's nothing
	// to do; return an error.
	if lowHash != nil && !s.cfg.DAG.BlockExists(lowHash) {
		return nil, &rpcmodel.RPCError{
			Code:    rpcmodel.ErrRPCBlockNotFound,
			Message: "Block not found in the DAG",
		}
	}

	// Retrieve the selected parent chain.
	removedChainHashes, addedChainHashes, err := s.cfg.DAG.SelectedParentChain(lowHash)
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
		return nil, &rpcmodel.RPCError{
			Code:    rpcmodel.ErrRPCInternal.Code,
			Message: fmt.Sprintf("could not collect chain blocks: %s", err),
		}
	}

	// Collect removedHashes.
	removedHashes := make([]string, len(removedChainHashes))
	for i, hash := range removedChainHashes {
		removedHashes[i] = hash.String()
	}

	result := &rpcmodel.GetChainFromBlockResult{
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
