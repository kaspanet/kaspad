package rpchandlers

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/subnetworkid"
)

// HandleGetBlock handles the respectively named RPC command
func HandleGetBlock(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	getBlockRequest := request.(*appmessage.GetBlockRequestMessage)

	// Load the raw block bytes from the database.
	hash, err := daghash.NewHashFromStr(getBlockRequest.Hash)
	if err != nil {
		errorMessage := &appmessage.GetBlockResponseMessage{}
		errorMessage.Error = &appmessage.RPCError{
			Message: fmt.Sprintf("Hash could not be parsed: %s", err),
		}
		return errorMessage, nil
	}
	if context.DAG.IsKnownInvalid(hash) {
		errorMessage := &appmessage.GetBlockResponseMessage{}
		errorMessage.Error = &appmessage.RPCError{
			Message: fmt.Sprintf("Block %s is known to be invalid", hash),
		}
		return errorMessage, nil
	}
	if context.DAG.IsKnownOrphan(hash) {
		errorMessage := &appmessage.GetBlockResponseMessage{}
		errorMessage.Error = &appmessage.RPCError{
			Message: fmt.Sprintf("Block %s is an orphan", hash),
		}
		return errorMessage, nil
	}
	block, err := context.DAG.BlockByHash(hash)
	if err != nil {
		errorMessage := &appmessage.GetBlockResponseMessage{}
		errorMessage.Error = &appmessage.RPCError{
			Message: fmt.Sprintf("Block %s not found", hash),
		}
		return errorMessage, nil
	}

	blockBytes, err := block.Bytes()
	if err != nil {
		return nil, err
	}

	// Handle partial blocks
	if getBlockRequest.SubnetworkID != "" {
		requestSubnetworkID, err := subnetworkid.NewFromStr(getBlockRequest.SubnetworkID)
		if err != nil {
			errorMessage := &appmessage.GetBlockResponseMessage{}
			errorMessage.Error = &appmessage.RPCError{
				Message: fmt.Sprintf("SubnetworkID could not be parsed: %s", err),
			}
			return errorMessage, nil
		}
		nodeSubnetworkID := context.Config.SubnetworkID

		if requestSubnetworkID != nil {
			if nodeSubnetworkID != nil {
				if !nodeSubnetworkID.IsEqual(requestSubnetworkID) {
					errorMessage := &appmessage.GetBlockResponseMessage{}
					errorMessage.Error = &appmessage.RPCError{
						Message: fmt.Sprintf("subnetwork %s does not match this partial node",
							getBlockRequest.SubnetworkID),
					}
					return errorMessage, nil
				}
				// nothing to do - partial node stores partial blocks
			} else {
				// Deserialize the block.
				msgBlock := block.MsgBlock()
				msgBlock.ConvertToPartial(requestSubnetworkID)
				var b bytes.Buffer
				err := msgBlock.Serialize(bufio.NewWriter(&b))
				if err != nil {
					return nil, err
				}
				blockBytes = b.Bytes()
			}
		}
	}

	response := appmessage.NewGetBlockResponseMessage()

	if getBlockRequest.IncludeBlockHex {
		response.BlockHex = hex.EncodeToString(blockBytes)
	}
	if getBlockRequest.IncludeBlockVerboseData {
		blockVerboseData, err := context.BuildBlockVerboseData(block, true)
		if err != nil {
			return nil, err
		}
		response.BlockVerboseData = blockVerboseData
	}

	return response, nil
}
