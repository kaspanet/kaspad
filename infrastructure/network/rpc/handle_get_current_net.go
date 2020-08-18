package rpc

// handleGetCurrentNet implements the getCurrentNet command.
func handleGetCurrentNet(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return s.dag.Params.Net, nil
}
