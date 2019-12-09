package rpc

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/kaspanet/kaspad/kaspajson"
	"github.com/kaspanet/kaspad/util/daghash"
	"strconv"
)

// handleGetBlockHeader implements the getBlockHeader command.
func handleGetBlockHeader(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*kaspajson.GetBlockHeaderCmd)

	// Fetch the header from chain.
	hash, err := daghash.NewHashFromStr(c.Hash)
	if err != nil {
		return nil, rpcDecodeHexError(c.Hash)
	}
	blockHeader, err := s.cfg.DAG.HeaderByHash(hash)
	if err != nil {
		return nil, &kaspajson.RPCError{
			Code:    kaspajson.ErrRPCBlockNotFound,
			Message: "Block not found",
		}
	}

	// When the verbose flag isn't set, simply return the serialized block
	// header as a hex-encoded string.
	if c.Verbose != nil && !*c.Verbose {
		var headerBuf bytes.Buffer
		err := blockHeader.Serialize(&headerBuf)
		if err != nil {
			context := "Failed to serialize block header"
			return nil, internalRPCError(err.Error(), context)
		}
		return hex.EncodeToString(headerBuf.Bytes()), nil
	}

	// The verbose flag is set, so generate the JSON object and return it.

	// Get the block chain height from chain.
	blockChainHeight, err := s.cfg.DAG.BlockChainHeightByHash(hash)
	if err != nil {
		context := "Failed to obtain block height"
		return nil, internalRPCError(err.Error(), context)
	}

	// Get the hashes for the next blocks unless there are none.
	var nextHashStrings []string
	if blockChainHeight < s.cfg.DAG.ChainHeight() { //TODO: (Ori) This is probably wrong. Done only for compilation
		childHashes, err := s.cfg.DAG.ChildHashesByHash(hash)
		if err != nil {
			context := "No next block"
			return nil, internalRPCError(err.Error(), context)
		}
		nextHashStrings = daghash.Strings(childHashes)
	}

	blockConfirmations, err := s.cfg.DAG.BlockConfirmationsByHash(hash)
	if err != nil {
		context := "Could not get block confirmations"
		return nil, internalRPCError(err.Error(), context)
	}

	params := s.cfg.DAGParams
	blockHeaderReply := kaspajson.GetBlockHeaderVerboseResult{
		Hash:                 c.Hash,
		Confirmations:        blockConfirmations,
		Height:               blockChainHeight,
		Version:              blockHeader.Version,
		VersionHex:           fmt.Sprintf("%08x", blockHeader.Version),
		HashMerkleRoot:       blockHeader.HashMerkleRoot.String(),
		AcceptedIDMerkleRoot: blockHeader.AcceptedIDMerkleRoot.String(),
		NextHashes:           nextHashStrings,
		ParentHashes:         daghash.Strings(blockHeader.ParentHashes),
		Nonce:                uint64(blockHeader.Nonce),
		Time:                 blockHeader.Timestamp.Unix(),
		Bits:                 strconv.FormatInt(int64(blockHeader.Bits), 16),
		Difficulty:           getDifficultyRatio(blockHeader.Bits, params),
	}
	return blockHeaderReply, nil
}
