package rpc

import (
	"github.com/kaspanet/kaspad/rpcmodel"
	"github.com/kaspanet/kaspad/util/network"
)

// handleAddManualNode handles addManualNode commands.
func handleAddManualNode(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*rpcmodel.AddManualNodeCmd)

	oneTry := c.OneTry != nil && *c.OneTry

	addr := network.NormalizeAddress(c.Addr, s.cfg.DAGParams.DefaultPort)
	var err error
	if oneTry {
		err = s.cfg.ConnMgr.Connect(addr, false)
	} else {
		err = s.cfg.ConnMgr.Connect(addr, true)
	}

	if err != nil {
		return nil, &rpcmodel.RPCError{
			Code:    rpcmodel.ErrRPCInvalidParameter,
			Message: err.Error(),
		}
	}

	// no data returned unless an error.
	return nil, nil
}
