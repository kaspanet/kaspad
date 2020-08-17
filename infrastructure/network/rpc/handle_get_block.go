package rpc

import (
	"bufio"
	"bytes"
	"encoding/hex"

	"github.com/kaspanet/kaspad/infrastructure/network/rpc/model"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/subnetworkid"
)

// handleGetBlock implements the getBlock command.
func handleGetBlock(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*model.GetBlockCmd)

	// Load the raw block bytes from the database.
	hash, err := daghash.NewHashFromStr(c.Hash)
	if err != nil {
		return nil, rpcDecodeHexError(c.Hash)
	}

	// Return an appropriate error if the block is known to be invalid
	if s.dag.IsKnownInvalid(hash) {
		return nil, &model.RPCError{
			Code:    model.ErrRPCBlockInvalid,
			Message: "Block is known to be invalid",
		}
	}

	// Return an appropriate error if the block is an orphan
	if s.dag.IsKnownOrphan(hash) {
		return nil, &model.RPCError{
			Code:    model.ErrRPCOrphanBlock,
			Message: "Block is an orphan",
		}
	}

	block, err := s.dag.BlockByHash(hash)
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

	// Handle partial blocks
	if c.Subnetwork != nil {
		requestSubnetworkID, err := subnetworkid.NewFromStr(*c.Subnetwork)
		if err != nil {
			return nil, &model.RPCError{
				Code:    model.ErrRPCInvalidRequest.Code,
				Message: "invalid subnetwork string",
			}
		}
		nodeSubnetworkID := s.cfg.SubnetworkID

		if requestSubnetworkID != nil {
			if nodeSubnetworkID != nil {
				if !nodeSubnetworkID.IsEqual(requestSubnetworkID) {
					return nil, &model.RPCError{
						Code:    model.ErrRPCInvalidRequest.Code,
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

	s.dag.RLock()
	defer s.dag.RUnlock()
	blockReply, err := buildGetBlockVerboseResult(s, block, c.VerboseTx == nil || !*c.VerboseTx)
	if err != nil {
		return nil, err
	}
	return blockReply, nil
}
