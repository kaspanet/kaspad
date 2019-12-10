package rpc

import (
	"github.com/kaspanet/kaspad/rpcmodel"
	"github.com/kaspanet/kaspad/util/network"
)

// handleRemoveManualNode handles removeManualNode command.
func handleRemoveManualNode(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*rpcmodel.RemoveManualNodeCmd)

	addr := network.NormalizeAddress(c.Addr, s.cfg.DAGParams.DefaultPort)
	err := s.cfg.ConnMgr.RemoveByAddr(addr)

	if err != nil {
		return nil, &rpcmodel.RPCError{
			Code:    rpcmodel.ErrRPCInvalidParameter,
			Message: err.Error(),
		}
	}

	// no data returned unless an error.
	return nil, nil
}
