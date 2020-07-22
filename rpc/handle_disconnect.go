package rpc

import (
	"github.com/kaspanet/kaspad/rpc/model"
	"github.com/kaspanet/kaspad/util/network"
)

// handleDisconnect handles disconnect commands.
func handleDisconnect(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*model.DisconnectCmd)

	address, err := network.NormalizeAddress(c.Address, s.dag.Params.DefaultPort)
	if err != nil {
		return nil, &model.RPCError{
			Code:    model.ErrRPCInvalidParameter,
			Message: err.Error(),
		}
	}

	s.connectionManager.RemoveConnection(address)
	return nil, nil
}
