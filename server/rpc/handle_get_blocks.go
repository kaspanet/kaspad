package rpc

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/database"
	"github.com/kaspanet/kaspad/kaspajson"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
)

const (
	// maxBlocksInGetBlocksResult is the max amount of blocks that are
	// allowed in a GetBlocksResult.
	maxBlocksInGetBlocksResult = 1000
)

func handleGetBlocks(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*kaspajson.GetBlocksCmd)
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

	// If startHash is not in the DAG, there's nothing to do; return an error.
	if startHash != nil && !s.cfg.DAG.HaveBlock(startHash) {
		return nil, &kaspajson.RPCError{
			Code:    kaspajson.ErrRPCBlockNotFound,
			Message: "Block not found",
		}
	}

	// Retrieve the block hashes.
	blockHashes, err := s.cfg.DAG.BlockHashesFrom(startHash, maxBlocksInGetBlocksResult)
	if err != nil {
		return nil, err
	}

	// Convert the hashes to strings
	hashes := make([]string, len(blockHashes))
	for i, blockHash := range blockHashes {
		hashes[i] = blockHash.String()
	}

	result := &kaspajson.GetBlocksResult{
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
	err := s.cfg.DB.View(func(dbTx database.Tx) error {
		for i, hash := range hashes {
			blockBytes, err := dbTx.FetchBlock(hash)
			if err != nil {
				return err
			}
			blocks[i] = blockBytes
		}
		return nil
	})
	if err != nil {
		return nil, err
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

func blockBytesToBlockVerboseResults(s *Server, blockBytesSlice [][]byte) ([]kaspajson.GetBlockVerboseResult, error) {
	verboseBlocks := make([]kaspajson.GetBlockVerboseResult, len(blockBytesSlice))
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
