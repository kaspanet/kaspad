package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/rpcmodel"
	"github.com/kaspanet/kaspad/util"
)

// handleSubmitBlock implements the submitBlock command.
func handleSubmitBlock(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*rpcmodel.SubmitBlockCmd)

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
		return nil, &rpcmodel.RPCError{
			Code:    rpcmodel.ErrRPCDeserialization,
			Message: "Block decode failed: " + err.Error(),
		}
	}

	// Process this block using the same rules as blocks coming from other
	// nodes. This will in turn relay it to the network like normal.
	_, err = s.cfg.SyncMgr.SubmitBlock(block, blockdag.BFNone)
	if err != nil {
		return nil, &rpcmodel.RPCError{
			Code:    rpcmodel.ErrRPCVerify,
			Message: fmt.Sprintf("Block rejected. Reason: %s", err),
		}
	}

	log.Infof("Accepted block %s via submitBlock", block.Hash())
	return nil, nil
}
