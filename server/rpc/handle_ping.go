package rpc

import (
	"github.com/daglabs/btcd/util/random"
	"github.com/daglabs/btcd/wire"
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
