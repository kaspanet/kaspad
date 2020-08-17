package rpc

// handleGetConnectionCount implements the getConnectionCount command.
func handleGetConnectionCount(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return s.connectionManager.ConnectionCount(), nil
}
