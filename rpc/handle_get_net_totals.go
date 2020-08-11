package rpc

import (
	"github.com/kaspanet/kaspad/rpc/model"
	"time"
)

// handleGetNetTotals implements the getNetTotals command.
func handleGetNetTotals(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	reply := &model.GetNetTotalsResult{
		TotalBytesRecv: 0,
		TotalBytesSent: 0,
		TimeMillis:     time.Now().UTC().UnixNano() / int64(time.Millisecond),
	}
	return reply, nil
}
