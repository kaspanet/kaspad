package rpc

import (
	"github.com/kaspanet/kaspad/btcjson"
	"github.com/kaspanet/kaspad/config"
)

// handleGetHashesPerSec implements the getHashesPerSec command.
func handleGetHashesPerSec(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	if config.ActiveConfig().SubnetworkID != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCInvalidRequest.Code,
			Message: "`getHashesPerSec` is not supported on partial nodes.",
		}
	}

	return int64(s.cfg.CPUMiner.HashesPerSecond()), nil
}
