package rpc

// handleGetBestBlockHash implements the getBestBlockHash command.
func handleGetBestBlockHash(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return s.cfg.DAG.SelectedTipHash().String(), nil
}
