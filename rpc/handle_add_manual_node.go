package rpc

import (
	"github.com/kaspanet/kaspad/rpcmodel"
	"github.com/kaspanet/kaspad/util/network"
)

// handleAddManualNode handles addManualNode commands.
func handleAddManualNode(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*rpcmodel.AddManualNodeCmd)

	oneTry := c.OneTry != nil && *c.OneTry

	addr, err := network.NormalizeAddress(c.Addr, s.dag.Params.DefaultPort)
	if err != nil {
		return nil, &rpcmodel.RPCError{
			Code:    rpcmodel.ErrRPCInvalidParameter,
			Message: err.Error(),
		}
	}

	s.connectionManager.AddConnectionRequest(addr, !oneTry)
	return nil, nil
}
