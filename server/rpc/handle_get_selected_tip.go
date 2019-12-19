package rpc

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/database"
	"github.com/kaspanet/kaspad/rpcmodel"
	"github.com/kaspanet/kaspad/util"
)

// handleGetSelectedTip implements the getSelectedTip command.
func handleGetSelectedTip(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	getSelectedTipCmd := cmd.(*rpcmodel.GetSelectedTipCmd)
	selectedTipHash := s.cfg.DAG.SelectedTipHash()

	var blockBytes []byte
	err := s.cfg.DB.View(func(dbTx database.Tx) error {
		var err error
		blockBytes, err = dbTx.FetchBlock(selectedTipHash)
		return err
	})
	if err != nil {
		return nil, &rpcmodel.RPCError{
			Code:    rpcmodel.ErrRPCBlockNotFound,
			Message: "Block not found",
		}
	}

	// When the verbose flag is set to false, simply return the serialized block
	// as a hex-encoded string (verbose flag is on by default).
	if getSelectedTipCmd.Verbose != nil && !*getSelectedTipCmd.Verbose {
		return hex.EncodeToString(blockBytes), nil
	}

	// Deserialize the block.
	blk, err := util.NewBlockFromBytes(blockBytes)
	if err != nil {
		context := "Failed to deserialize block"
		return nil, internalRPCError(err.Error(), context)
	}

	blockVerboseResult, err := buildGetBlockVerboseResult(s, blk, getSelectedTipCmd.VerboseTx == nil || !*getSelectedTipCmd.VerboseTx)
	if err != nil {
		return nil, err
	}
	return blockVerboseResult, nil
}
