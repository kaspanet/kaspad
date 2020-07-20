package rpc

// handleGetSelectedTipHash implements the getSelectedTipHash command.
func handleGetSelectedTipHash(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return s.dag.SelectedTipHash().String(), nil
}
