package rpc

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/infrastructure/network/rpc/model"
)

// handleGetSelectedTip implements the getSelectedTip command.
func handleGetSelectedTip(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	getSelectedTipCmd := cmd.(*model.GetSelectedTipCmd)
	selectedTipHash := s.dag.SelectedTipHash()

	block, err := s.dag.BlockByHash(selectedTipHash)
	if err != nil {
		return nil, &model.RPCError{
			Code:    model.ErrRPCBlockNotFound,
			Message: "Block not found",
		}
	}
	blockBytes, err := block.Bytes()
	if err != nil {
		return nil, &model.RPCError{
			Code:    model.ErrRPCBlockInvalid,
			Message: "Cannot serialize block",
		}
	}

	// When the verbose flag is set to false, simply return the serialized block
	// as a hex-encoded string (verbose flag is on by default).
	if getSelectedTipCmd.Verbose != nil && !*getSelectedTipCmd.Verbose {
		return hex.EncodeToString(blockBytes), nil
	}

	blockVerboseResult, err := buildGetBlockVerboseResult(s, block, getSelectedTipCmd.VerboseTx == nil || !*getSelectedTipCmd.VerboseTx)
	if err != nil {
		return nil, err
	}
	return blockVerboseResult, nil
}
