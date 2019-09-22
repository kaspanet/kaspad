package rpc

// handleGetBlockHash implements the getBlockHash command.
// This command had been (possibly temporarily) dropped.
// Originally it relied on height, which no longer makes sense.
func handleGetBlockHash(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return nil, ErrRPCUnimplemented
}
