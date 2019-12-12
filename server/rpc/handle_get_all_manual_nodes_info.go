package rpc

import "github.com/kaspanet/kaspad/rpcmodel"

// handleGetAllManualNodesInfo handles getAllManualNodesInfo commands.
func handleGetAllManualNodesInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*rpcmodel.GetAllManualNodesInfoCmd)
	return getManualNodesInfo(s, c.Details, "")
}
