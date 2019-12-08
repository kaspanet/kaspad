package rpc

import "github.com/kaspanet/kaspad/btcjson"

// handleGetAllManualNodesInfo handles getAllManualNodesInfo commands.
func handleGetAllManualNodesInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.GetAllManualNodesInfoCmd)
	return getManualNodesInfo(s, c.Details, "")
}
