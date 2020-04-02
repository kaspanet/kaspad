package rpc

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/rpcmodel"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/subnetworkid"
)

// handleGetBlock implements the getBlock command.
func handleGetBlock(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*rpcmodel.GetBlockCmd)

	// Load the raw block bytes from the database.
	hash, err := daghash.NewHashFromStr(c.Hash)
	if err != nil {
		return nil, rpcDecodeHexError(c.Hash)
	}

	// Return an appropriate error if the block is known to be invalid
	if s.cfg.DAG.IsKnownInvalid(hash) {
		return nil, &rpcmodel.RPCError{
			Code:    rpcmodel.ErrRPCBlockInvalid,
			Message: "Block is known to be invalid",
		}
	}

	// Return an appropriate error if the block is an orphan
	if s.cfg.DAG.IsKnownOrphan(hash) {
		return nil, &rpcmodel.RPCError{
			Code:    rpcmodel.ErrRPCOrphanBlock,
			Message: "Block is an orphan",
		}
	}

	block, err := s.cfg.DAG.BlockByHash(hash)
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

	// Handle partial blocks
	if c.Subnetwork != nil {
		requestSubnetworkID, err := subnetworkid.NewFromStr(*c.Subnetwork)
		if err != nil {
			return nil, &rpcmodel.RPCError{
				Code:    rpcmodel.ErrRPCInvalidRequest.Code,
				Message: "invalid subnetwork string",
			}
		}
		nodeSubnetworkID := config.ActiveConfig().SubnetworkID

		if requestSubnetworkID != nil {
			if nodeSubnetworkID != nil {
				if !nodeSubnetworkID.IsEqual(requestSubnetworkID) {
					return nil, &rpcmodel.RPCError{
						Code:    rpcmodel.ErrRPCInvalidRequest.Code,
						Message: "subnetwork does not match this partial node",
					}
				}
				// nothing to do - partial node stores partial blocks
			} else {
				// Deserialize the block.
				msgBlock := block.MsgBlock()
				msgBlock.ConvertToPartial(requestSubnetworkID)
				var b bytes.Buffer
				msgBlock.Serialize(bufio.NewWriter(&b))
				blockBytes = b.Bytes()
			}
		}
	}

	// When the verbose flag is set to false, simply return the serialized block
	// as a hex-encoded string (verbose flag is on by default).
	if c.Verbose != nil && !*c.Verbose {
		return hex.EncodeToString(blockBytes), nil
	}

	// The verbose flag is set, so generate the JSON object and return it.

	// Deserialize the block.
	block, err = util.NewBlockFromBytes(blockBytes)
	if err != nil {
		context := "Failed to deserialize block"
		return nil, internalRPCError(err.Error(), context)
	}

	s.cfg.DAG.RLock()
	defer s.cfg.DAG.RUnlock()
	blockReply, err := buildGetBlockVerboseResult(s, block, c.VerboseTx == nil || !*c.VerboseTx)
	if err != nil {
		return nil, err
	}
	return blockReply, nil
}
