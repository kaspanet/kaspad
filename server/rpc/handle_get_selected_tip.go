package rpc

import (
	"encoding/hex"
	"github.com/daglabs/btcd/btcjson"
	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/util"
)

// handleGetSelectedTip implements the getSelectedTip command.
func handleGetSelectedTip(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.GetSelectedTipCmd)
	hash := s.cfg.DAG.SelectedTipHash()

	var blkBytes []byte
	err := s.cfg.DB.View(func(dbTx database.Tx) error {
		var err error
		blkBytes, err = dbTx.FetchBlock(hash)
		return err
	})
	if err != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCBlockNotFound,
			Message: "Block not found",
		}
	}

	// When the verbose flag isn't set, simply return the serialized block
	// as a hex-encoded string.
	if c.Verbose != nil && !*c.Verbose {
		return hex.EncodeToString(blkBytes), nil
	}

	// Deserialize the block.
	blk, err := util.NewBlockFromBytes(blkBytes)
	if err != nil {
		context := "Failed to deserialize block"
		return nil, internalRPCError(err.Error(), context)
	}

	blockReply, err := buildGetBlockVerboseResult(s, blk, c.VerboseTx == nil || !*c.VerboseTx)
	if err != nil {
		return nil, err
	}
	return blockReply, nil
}
