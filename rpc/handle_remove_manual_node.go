package rpc

import (
	"github.com/kaspanet/kaspad/rpc/model"
	"github.com/kaspanet/kaspad/util/network"
)

// handleRemoveManualNode handles removeManualNode command.
func handleRemoveManualNode(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*model.RemoveManualNodeCmd)

	addr, err := network.NormalizeAddress(c.Addr, s.dag.Params.DefaultPort)
	if err != nil {
		return nil, &model.RPCError{
			Code:    model.ErrRPCInvalidParameter,
			Message: err.Error(),
		}
	}

	err = s.connectionManager.RemoveByAddr(addr)
	if err != nil {
		return nil, &model.RPCError{
			Code:    model.ErrRPCInvalidParameter,
			Message: err.Error(),
		}
	}

	// no data returned unless an error.
	return nil, nil
}
