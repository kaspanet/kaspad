package rpc

import (
	"github.com/kaspanet/kaspad/kaspajson"
	"github.com/kaspanet/kaspad/util/network"
)

// handleRemoveManualNode handles removeManualNode command.
func handleRemoveManualNode(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*kaspajson.RemoveManualNodeCmd)

	addr := network.NormalizeAddress(c.Addr, s.cfg.DAGParams.DefaultPort)
	err := s.cfg.ConnMgr.RemoveByAddr(addr)

	if err != nil {
		return nil, &kaspajson.RPCError{
			Code:    kaspajson.ErrRPCInvalidParameter,
			Message: err.Error(),
		}
	}

	// no data returned unless an error.
	return nil, nil
}
