package rpc

import (
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/kaspajson"
)

// handleGetNetworkHashPS implements the getNetworkHashPs command.
// This command had been (possibly temporarily) dropped.
// Originally it relied on height, which no longer makes sense.
func handleGetNetworkHashPS(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	if config.ActiveConfig().SubnetworkID != nil {
		return nil, &kaspajson.RPCError{
			Code:    kaspajson.ErrRPCInvalidRequest.Code,
			Message: "`getNetworkHashPS` is not supported on partial nodes.",
		}
	}

	return nil, ErrRPCUnimplemented
}
