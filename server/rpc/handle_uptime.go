package rpc

import (
	"github.com/kaspanet/kaspad/util/mstime"
	"time"
)

// handleUptime implements the uptime command.
func handleUptime(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return mstime.TimeToUnixMilli(time.Now()) - s.cfg.StartupTime, nil
}
