package rpc

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"github.com/daglabs/btcd/btcjson"
	"github.com/daglabs/btcd/config"
	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/util/subnetworkid"
	"github.com/daglabs/btcd/wire"
)

// handleGetBlock implements the getBlock command.
func handleGetBlock(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.GetBlockCmd)

	// Load the raw block bytes from the database.
	hash, err := daghash.NewHashFromStr(c.Hash)
	if err != nil {
		return nil, rpcDecodeHexError(c.Hash)
	}

	// Return an appropriate error if the block is an orphan
	if s.cfg.DAG.IsKnownOrphan(hash) {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCOrphanBlock,
			Message: "Block is an orphan",
		}
	}

	var blkBytes []byte
	err = s.cfg.DB.View(func(dbTx database.Tx) error {
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

	// Handle partial blocks
	if c.Subnetwork != nil {
		requestSubnetworkID, err := subnetworkid.NewFromStr(*c.Subnetwork)
		if err != nil {
			return nil, &btcjson.RPCError{
				Code:    btcjson.ErrRPCInvalidRequest.Code,
				Message: "invalid subnetwork string",
			}
		}
		nodeSubnetworkID := config.ActiveConfig().SubnetworkID

		if requestSubnetworkID != nil {
			if nodeSubnetworkID != nil {
				if !nodeSubnetworkID.IsEqual(requestSubnetworkID) {
					return nil, &btcjson.RPCError{
						Code:    btcjson.ErrRPCInvalidRequest.Code,
						Message: "subnetwork does not match this partial node",
					}
				}
				// nothing to do - partial node stores partial blocks
			} else {
				// Deserialize the block.
				var msgBlock wire.MsgBlock
				err = msgBlock.Deserialize(bytes.NewReader(blkBytes))
				if err != nil {
					context := "Failed to deserialize block"
					return nil, internalRPCError(err.Error(), context)
				}
				msgBlock.ConvertToPartial(requestSubnetworkID)
				var b bytes.Buffer
				msgBlock.Serialize(bufio.NewWriter(&b))
				blkBytes = b.Bytes()
			}
		}
	}

	// When the verbose flag is set to false, simply return the serialized block
	// as a hex-encoded string (verbose flag is on by default).
	if c.Verbose != nil && !*c.Verbose {
		return hex.EncodeToString(blkBytes), nil
	}

	// The verbose flag is set, so generate the JSON object and return it.

	// Deserialize the block.
	blk, err := util.NewBlockFromBytes(blkBytes)
	if err != nil {
		context := "Failed to deserialize block"
		return nil, internalRPCError(err.Error(), context)
	}

	s.cfg.DAG.RLock()
	defer s.cfg.DAG.RUnlock()
	blockReply, err := buildGetBlockVerboseResult(s, blk, c.VerboseTx == nil || !*c.VerboseTx)
	if err != nil {
		return nil, err
	}
	return blockReply, nil
}
