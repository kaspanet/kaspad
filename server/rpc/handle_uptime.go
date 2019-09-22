package rpc

import "time"

// handleUptime implements the uptime command.
func handleUptime(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return time.Now().Unix() - s.cfg.StartupTime, nil
}
