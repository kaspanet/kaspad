package rpc

import "github.com/kaspanet/kaspad/jsonrpc"

// handleGetAllManualNodesInfo handles getAllManualNodesInfo commands.
func handleGetAllManualNodesInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*jsonrpc.GetAllManualNodesInfoCmd)
	return getManualNodesInfo(s, c.Details, "")
}
