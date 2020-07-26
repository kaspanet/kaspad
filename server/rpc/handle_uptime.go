package rpc

import (
	"github.com/kaspanet/kaspad/util/mstime"
)

// handleUptime implements the uptime command.
func handleUptime(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return mstime.Now().UnixMilliseconds() - s.cfg.StartupTime, nil
}
