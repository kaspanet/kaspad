package rpc

// handleStop implements the stop command.
func handleStop(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	select {
	case s.requestProcessShutdown <- struct{}{}:
	default:
	}
	return "kaspad stopping.", nil
}
