package rpc

import (
	"fmt"
	"github.com/kaspanet/kaspad/rpcmodel"
	"github.com/kaspanet/kaspad/util/daghash"
)

// handleRescanBlocks implements the rescanBlocks command extension for
// websocket connections.
//
// NOTE: This extension is ported from github.com/decred/dcrd
func handleRescanBlocks(wsc *wsClient, icmd interface{}) (interface{}, error) {
	cmd, ok := icmd.(*rpcmodel.RescanBlocksCmd)
	if !ok {
		return nil, rpcmodel.ErrRPCInternal
	}

	// Load client's transaction filter. Must exist in order to continue.
	filter := wsc.FilterData()
	if filter == nil {
		return nil, &rpcmodel.RPCError{
			Code:    rpcmodel.ErrRPCMisc,
			Message: "Transaction filter must be loaded before rescanning",
		}
	}

	blockHashes := make([]*daghash.Hash, len(cmd.BlockHashes))

	for i := range cmd.BlockHashes {
		hash, err := daghash.NewHashFromStr(cmd.BlockHashes[i])
		if err != nil {
			return nil, err
		}
		blockHashes[i] = hash
	}

	discoveredData := make([]rpcmodel.RescannedBlock, 0, len(blockHashes))

	// Iterate over each block in the request and rescan. When a block
	// contains relevant transactions, add it to the response.
	bc := wsc.server.dag
	params := wsc.server.dag.Params
	var lastBlockHash *daghash.Hash
	for i := range blockHashes {
		block, err := bc.BlockByHash(blockHashes[i])
		if err != nil {
			return nil, &rpcmodel.RPCError{
				Code:    rpcmodel.ErrRPCBlockNotFound,
				Message: "Failed to fetch block: " + err.Error(),
			}
		}
		if lastBlockHash != nil && !block.MsgBlock().Header.ParentHashes[0].IsEqual(lastBlockHash) { // TODO: (Stas) This is likely wrong. Modified to satisfy compilation.
			return nil, &rpcmodel.RPCError{
				Code: rpcmodel.ErrRPCInvalidParameter,
				Message: fmt.Sprintf("Block %s is not a child of %s",
					blockHashes[i], lastBlockHash),
			}
		}
		lastBlockHash = blockHashes[i]

		transactions := rescanBlockFilter(filter, block, params)
		if len(transactions) != 0 {
			discoveredData = append(discoveredData, rpcmodel.RescannedBlock{
				Hash:         cmd.BlockHashes[i],
				Transactions: transactions,
			})
		}
	}

	return &discoveredData, nil
}
