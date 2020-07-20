package rpc

import "github.com/kaspanet/kaspad/rpc/model"

// handleGetAllManualNodesInfo handles getAllManualNodesInfo commands.
func handleGetAllManualNodesInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*model.GetAllManualNodesInfoCmd)
	return getManualNodesInfo(s, c.Details, "")
}
