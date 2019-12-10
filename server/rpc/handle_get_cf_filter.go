package rpc

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/rpcmodel"
	"github.com/kaspanet/kaspad/util/daghash"
)

// handleGetCFilter implements the getCFilter command.
func handleGetCFilter(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	if s.cfg.CfIndex == nil {
		return nil, &rpcmodel.RPCError{
			Code:    rpcmodel.ErrRPCNoCFIndex,
			Message: "The CF index must be enabled for this command",
		}
	}

	c := cmd.(*rpcmodel.GetCFilterCmd)
	hash, err := daghash.NewHashFromStr(c.Hash)
	if err != nil {
		return nil, rpcDecodeHexError(c.Hash)
	}

	filterBytes, err := s.cfg.CfIndex.FilterByBlockHash(hash, c.FilterType)
	if err != nil {
		log.Debugf("Could not find committed filter for %s: %s",
			hash, err)
		return nil, &rpcmodel.RPCError{
			Code:    rpcmodel.ErrRPCBlockNotFound,
			Message: "Block not found",
		}
	}

	log.Debugf("Found committed filter for %s", hash)
	return hex.EncodeToString(filterBytes), nil
}
