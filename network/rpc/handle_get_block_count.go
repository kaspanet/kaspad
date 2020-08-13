package rpc

// handleGetBlockCount implements the getBlockCount command.
func handleGetBlockCount(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return s.dag.BlockCount(), nil
}
