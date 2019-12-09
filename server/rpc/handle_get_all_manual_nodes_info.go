package rpc

import "github.com/kaspanet/kaspad/kaspajson"

// handleGetAllManualNodesInfo handles getAllManualNodesInfo commands.
func handleGetAllManualNodesInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*kaspajson.GetAllManualNodesInfoCmd)
	return getManualNodesInfo(s, c.Details, "")
}
