package rpc

import (
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/rpcmodel"
	"github.com/kaspanet/kaspad/version"
)

// handleGetInfo implements the getInfo command. We only return the fields
// that are not related to wallet functionality.
func handleGetInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	ret := &rpcmodel.InfoDAGResult{
		Version:         version.Version(),
		ProtocolVersion: int32(maxProtocolVersion),
		Blocks:          s.cfg.DAG.BlockCount(),
		TimeOffset:      int64(s.cfg.TimeSource.Offset().Seconds()),
		Connections:     s.cfg.ConnMgr.ConnectedCount(),
		Proxy:           config.ActiveConfig().Proxy,
		Difficulty:      getDifficultyRatio(s.cfg.DAG.CurrentBits(), s.cfg.DAGParams),
		TestNet:         config.ActiveConfig().TestNet,
		DevNet:          config.ActiveConfig().DevNet,
		RelayFee:        config.ActiveConfig().MinRelayTxFee.ToKAS(),
	}

	return ret, nil
}
