package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/kaspanet/kaspad/domain/blockdag"
	"github.com/kaspanet/kaspad/network/rpc/model"
	"github.com/kaspanet/kaspad/util"
)

// handleSubmitBlock implements the submitBlock command.
func handleSubmitBlock(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*model.SubmitBlockCmd)

	// Deserialize the submitted block.
	hexStr := c.HexBlock
	serializedBlock, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, rpcDecodeHexError(hexStr)
	}

	block, err := util.NewBlockFromBytes(serializedBlock)
	if err != nil {
		return nil, &model.RPCError{
			Code:    model.ErrRPCDeserialization,
			Message: "Block decode failed: " + err.Error(),
		}
	}

	err = s.protocolManager.AddBlock(block, blockdag.BFDisallowDelay|blockdag.BFDisallowOrphans)
	if err != nil {
		return nil, &model.RPCError{
			Code:    model.ErrRPCVerify,
			Message: fmt.Sprintf("Block rejected. Reason: %s", err),
		}
	}

	log.Infof("Accepted block %s via submitBlock", block.Hash())
	return nil, nil
}
