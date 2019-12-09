package rpc

import (
	"github.com/kaspanet/kaspad/jsonrpc"
	"github.com/kaspanet/kaspad/util/daghash"
)

// handleGetCFilterHeader implements the getCFilterHeader command.
func handleGetCFilterHeader(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	if s.cfg.CfIndex == nil {
		return nil, &jsonrpc.RPCError{
			Code:    jsonrpc.ErrRPCNoCFIndex,
			Message: "The CF index must be enabled for this command",
		}
	}

	c := cmd.(*jsonrpc.GetCFilterHeaderCmd)
	hash, err := daghash.NewHashFromStr(c.Hash)
	if err != nil {
		return nil, rpcDecodeHexError(c.Hash)
	}

	headerBytes, err := s.cfg.CfIndex.FilterHeaderByBlockHash(hash, c.FilterType)
	if len(headerBytes) > 0 {
		log.Debugf("Found header of committed filter for %s", hash)
	} else {
		log.Debugf("Could not find header of committed filter for %s: %s",
			hash, err)
		return nil, &jsonrpc.RPCError{
			Code:    jsonrpc.ErrRPCBlockNotFound,
			Message: "Block not found",
		}
	}

	hash.SetBytes(headerBytes)
	return hash.String(), nil
}
