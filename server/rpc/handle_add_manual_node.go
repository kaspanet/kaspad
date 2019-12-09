package rpc

import (
	"github.com/kaspanet/kaspad/jsonrpc"
	"github.com/kaspanet/kaspad/util/network"
)

// handleAddManualNode handles addManualNode commands.
func handleAddManualNode(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*jsonrpc.AddManualNodeCmd)

	oneTry := c.OneTry != nil && *c.OneTry

	addr := network.NormalizeAddress(c.Addr, s.cfg.DAGParams.DefaultPort)
	var err error
	if oneTry {
		err = s.cfg.ConnMgr.Connect(addr, false)
	} else {
		err = s.cfg.ConnMgr.Connect(addr, true)
	}

	if err != nil {
		return nil, &jsonrpc.RPCError{
			Code:    jsonrpc.ErrRPCInvalidParameter,
			Message: err.Error(),
		}
	}

	// no data returned unless an error.
	return nil, nil
}
