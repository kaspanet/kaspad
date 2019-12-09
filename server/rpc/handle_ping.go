package rpc

import (
	"github.com/kaspanet/kaspad/util/random"
	"github.com/kaspanet/kaspad/wire"
)

// handlePing implements the ping command.
func handlePing(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	// Ask server to ping \o_
	nonce, err := random.Uint64()
	if err != nil {
		return nil, internalRPCError("Not sending ping - failed to "+
			"generate nonce: "+err.Error(), "")
	}
	s.cfg.ConnMgr.BroadcastMessage(wire.NewMsgPing(nonce))

	return nil, nil
}
