package rpc

import (
	"github.com/kaspanet/kaspad/rpc/model"
	"time"
)

// handleGetNetTotals implements the getNetTotals command.
func handleGetNetTotals(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	totalBytesRecv, totalBytesSent := s.connectionManager.NetTotals()
	reply := &model.GetNetTotalsResult{
		TotalBytesRecv: totalBytesRecv,
		TotalBytesSent: totalBytesSent,
		TimeMillis:     time.Now().UTC().UnixNano() / int64(time.Millisecond),
	}
	return reply, nil
}
