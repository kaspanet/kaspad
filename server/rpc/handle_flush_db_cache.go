package rpc

// handleFlushDBCache flushes the db cache to the disk.
// TODO: (Ori) This is a temporary function for dev use. It needs to be removed.
func handleFlushDBCache(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	err := s.cfg.DB.FlushCache()
	return nil, err
}
