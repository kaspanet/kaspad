package rpc

// handleGetGenerate implements the getGenerate command.
func handleGetGenerate(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return s.cfg.CPUMiner.IsMining(), nil
}
