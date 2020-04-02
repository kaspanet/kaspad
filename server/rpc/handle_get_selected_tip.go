package rpc

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/rpcmodel"
)

// handleGetSelectedTip implements the getSelectedTip command.
func handleGetSelectedTip(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	getSelectedTipCmd := cmd.(*rpcmodel.GetSelectedTipCmd)
	selectedTipHash := s.cfg.DAG.SelectedTipHash()

	block, err := s.cfg.DAG.BlockByHash(selectedTipHash)
	if err != nil {
		return nil, &rpcmodel.RPCError{
			Code:    rpcmodel.ErrRPCBlockNotFound,
			Message: "Block not found",
		}
	}
	blockBytes, err := block.Bytes()
	if err != nil {
		return nil, &rpcmodel.RPCError{
			Code:    rpcmodel.ErrRPCBlockInvalid,
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
