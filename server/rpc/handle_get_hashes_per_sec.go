package rpc

import (
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/rpcmodel"
)

// handleGetHashesPerSec implements the getHashesPerSec command.
func handleGetHashesPerSec(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	if config.ActiveConfig().SubnetworkID != nil {
		return nil, &rpcmodel.RPCError{
			Code:    rpcmodel.ErrRPCInvalidRequest.Code,
			Message: "`getHashesPerSec` is not supported on partial nodes.",
		}
	}

	return int64(s.cfg.CPUMiner.HashesPerSecond()), nil
}
