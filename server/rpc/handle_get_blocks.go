package rpc

import (
	"encoding/hex"
	"github.com/daglabs/btcd/btcjson"
	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/util/daghash"
)

const (
	// maxBlocksInGetBlocksResult is the max amount of blocks that are
	// allowed in a GetBlocksResult.
	maxBlocksInGetBlocksResult = 1000
)

func handleGetBlocks(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.GetBlocksCmd)
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
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCBlockNotFound,
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

	result := &btcjson.GetBlocksResult{
		Hashes: hashes,
		Blocks: nil,
	}

	// If the user specified to include the blocks, collect them as well.
	if c.IncludeBlocks {
		if c.VerboseBlocks {
			getBlockVerboseResults, err := hashesToGetBlockVerboseResults(s, blockHashes)
			if err != nil {
				return nil, err
			}
			result.RawBlocks = getBlockVerboseResults
		} else {
			blocks, err := hashesToBlockStrings(s, blockHashes)
			if err != nil {
				return nil, err
			}
			result.Blocks = blocks
		}
	}

	return result, nil
}

func hashesToBlockStrings(s *Server, hashes []*daghash.Hash) ([]string, error) {
	blocks := make([]string, len(hashes))
	err := s.cfg.DB.View(func(dbTx database.Tx) error {
		for i, hash := range hashes {
			blockBytes, err := dbTx.FetchBlock(hash)
			if err != nil {
				return err
			}
			blocks[i] = hex.EncodeToString(blockBytes)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return blocks, nil
}
