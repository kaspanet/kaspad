package rpc

import (
	"github.com/daglabs/btcd/btcjson"
	"github.com/daglabs/btcd/config"
	"github.com/daglabs/btcd/version"
)

// handleGetInfo implements the getInfo command. We only return the fields
// that are not related to wallet functionality.
func handleGetInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	ret := &btcjson.InfoDAGResult{
		Version:         int32(1000000*version.AppMajor + 10000*version.AppMinor + 100*version.AppPatch),
		ProtocolVersion: int32(maxProtocolVersion),
		Blocks:          s.cfg.DAG.BlockCount(),
		TimeOffset:      int64(s.cfg.TimeSource.Offset().Seconds()),
		Connections:     s.cfg.ConnMgr.ConnectedCount(),
		Proxy:           config.MainConfig().Proxy,
		Difficulty:      getDifficultyRatio(s.cfg.DAG.CurrentBits(), s.cfg.DAGParams),
		TestNet:         config.MainConfig().TestNet,
		DevNet:          config.MainConfig().DevNet,
		RelayFee:        config.MainConfig().MinRelayTxFee.ToBTC(),
	}

	return ret, nil
}
