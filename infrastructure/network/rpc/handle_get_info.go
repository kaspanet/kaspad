package rpc

import (
	"github.com/kaspanet/kaspad/infrastructure/network/rpc/model"
	"github.com/kaspanet/kaspad/version"
)

// handleGetInfo implements the getInfo command. We only return the fields
// that are not related to wallet functionality.
func handleGetInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	ret := &model.InfoDAGResult{
		Version:         version.Version(),
		ProtocolVersion: int32(maxProtocolVersion),
		Blocks:          s.dag.BlockCount(),
		Connections:     int32(s.connectionManager.ConnectionCount()),
		Proxy:           s.cfg.Proxy,
		Difficulty:      getDifficultyRatio(s.dag.CurrentBits(), s.dag.Params),
		Testnet:         s.cfg.Testnet,
		Devnet:          s.cfg.Devnet,
		RelayFee:        s.cfg.MinRelayTxFee.ToKAS(),
	}

	return ret, nil
}
