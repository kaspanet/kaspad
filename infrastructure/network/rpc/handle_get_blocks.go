package rpc

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/infrastructure/network/rpc/model"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
)

const (
	// maxBlocksInGetBlocksResult is the max amount of blocks that are
	// allowed in a GetBlocksResult.
	maxBlocksInGetBlocksResult = 1000
)

func handleGetBlocks(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*model.GetBlocksCmd)
	var lowHash *daghash.Hash
	if c.LowHash != nil {
		lowHash = &daghash.Hash{}
		err := daghash.Decode(lowHash, *c.LowHash)
		if err != nil {
			return nil, rpcDecodeHexError(*c.LowHash)
		}
	}

	s.dag.RLock()
	defer s.dag.RUnlock()

	// If lowHash is not in the DAG, there's nothing to do; return an error.
	if lowHash != nil && !s.dag.IsKnownBlock(lowHash) {
		return nil, &model.RPCError{
			Code:    model.ErrRPCBlockNotFound,
			Message: "Block not found",
		}
	}

	// Retrieve the block hashes.
	blockHashes, err := s.dag.BlockHashesFrom(lowHash, maxBlocksInGetBlocksResult)
	if err != nil {
		return nil, err
	}

	// Convert the hashes to strings
	hashes := make([]string, len(blockHashes))
	for i, blockHash := range blockHashes {
		hashes[i] = blockHash.String()
	}

	result := &model.GetBlocksResult{
		Hashes:        hashes,
		RawBlocks:     nil,
		VerboseBlocks: nil,
	}

	// Include more data if requested
	if c.IncludeRawBlockData || c.IncludeVerboseBlockData {
		blockBytesSlice, err := hashesToBlockBytes(s, blockHashes)
		if err != nil {
			return nil, err
		}
		if c.IncludeRawBlockData {
			result.RawBlocks = blockBytesToStrings(blockBytesSlice)
		}
		if c.IncludeVerboseBlockData {
			verboseBlocks, err := blockBytesToBlockVerboseResults(s, blockBytesSlice)
			if err != nil {
				return nil, err
			}
			result.VerboseBlocks = verboseBlocks
		}
	}

	return result, nil
}

func hashesToBlockBytes(s *Server, hashes []*daghash.Hash) ([][]byte, error) {
	blocks := make([][]byte, len(hashes))
	for i, hash := range hashes {
		block, err := s.dag.BlockByHash(hash)
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

func blockBytesToBlockVerboseResults(s *Server, blockBytesSlice [][]byte) ([]model.GetBlockVerboseResult, error) {
	verboseBlocks := make([]model.GetBlockVerboseResult, len(blockBytesSlice))
	for i, blockBytes := range blockBytesSlice {
		block, err := util.NewBlockFromBytes(blockBytes)
		if err != nil {
			return nil, err
		}
		getBlockVerboseResult, err := buildGetBlockVerboseResult(s, block, false)
		if err != nil {
			return nil, err
		}
		verboseBlocks[i] = *getBlockVerboseResult
	}
	return verboseBlocks, nil
}
