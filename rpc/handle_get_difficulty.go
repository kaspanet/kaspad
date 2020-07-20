package rpc

// handleGetDifficulty implements the getDifficulty command.
func handleGetDifficulty(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return getDifficultyRatio(s.cfg.DAG.SelectedTipHeader().Bits, s.cfg.DAG.Params), nil
}
