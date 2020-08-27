package rpc

// handleGetFinalityConflicts implements the getFinalityConflicts command.
func handleGetFinalityConflicts(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return s.dag.FinalityConflicts(), nil
}
