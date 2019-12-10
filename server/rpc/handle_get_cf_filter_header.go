package rpc

import (
	"github.com/kaspanet/kaspad/rpcmodel"
	"github.com/kaspanet/kaspad/util/daghash"
)

// handleGetCFilterHeader implements the getCFilterHeader command.
func handleGetCFilterHeader(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	if s.cfg.CfIndex == nil {
		return nil, &rpcmodel.RPCError{
			Code:    rpcmodel.ErrRPCNoCFIndex,
			Message: "The CF index must be enabled for this command",
		}
	}

	c := cmd.(*rpcmodel.GetCFilterHeaderCmd)
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
		return nil, &rpcmodel.RPCError{
			Code:    rpcmodel.ErrRPCBlockNotFound,
			Message: "Block not found",
		}
	}

	hash.SetBytes(headerBytes)
	return hash.String(), nil
}
