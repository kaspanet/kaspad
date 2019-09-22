package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/btcjson"
	"github.com/daglabs/btcd/util"
)

// handleSubmitBlock implements the submitBlock command.
func handleSubmitBlock(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.SubmitBlockCmd)

	// Deserialize the submitted block.
	hexStr := c.HexBlock
	if len(hexStr)%2 != 0 {
		hexStr = "0" + c.HexBlock
	}
	serializedBlock, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, rpcDecodeHexError(hexStr)
	}

	block, err := util.NewBlockFromBytes(serializedBlock)
	if err != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCDeserialization,
			Message: "Block decode failed: " + err.Error(),
		}
	}

	// Process this block using the same rules as blocks coming from other
	// nodes.  This will in turn relay it to the network like normal.
	_, err = s.cfg.SyncMgr.SubmitBlock(block, blockdag.BFNone)
	if err != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCVerify,
			Message: fmt.Sprintf("Block rejected. Reason: %s", err),
		}
	}

	log.Infof("Accepted block %s via submitBlock", block.Hash())
	return nil, nil
}
