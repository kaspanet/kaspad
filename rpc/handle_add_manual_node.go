package rpc

import (
	"github.com/kaspanet/kaspad/rpc/model"
	"github.com/kaspanet/kaspad/util/network"
)

// handleAddManualNode handles addManualNode commands.
func handleAddManualNode(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*model.AddManualNodeCmd)

	oneTry := c.OneTry != nil && *c.OneTry

	addr, err := network.NormalizeAddress(c.Addr, s.dag.Params.DefaultPort)
	if err != nil {
		return nil, &model.RPCError{
			Code:    model.ErrRPCInvalidParameter,
			Message: err.Error(),
		}
	}

	s.connectionManager.AddConnectionRequest(addr, !oneTry)
	return nil, nil
}
