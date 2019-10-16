package rpc

import (
	"github.com/daglabs/btcd/btcjson"
	"github.com/daglabs/btcd/config"
)

// handleGetMiningInfo implements the getMiningInfo command. We only return the
// fields that are not related to wallet functionality.
func handleGetMiningInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	if config.MainConfig().SubnetworkID != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCInvalidRequest.Code,
			Message: "`getMiningInfo` is not supported on partial nodes.",
		}
	}

	// Create a default getNetworkHashPs command to use defaults and make
	// use of the existing getNetworkHashPs handler.
	gnhpsCmd := btcjson.NewGetNetworkHashPSCmd(nil, nil)
	networkHashesPerSecIface, err := handleGetNetworkHashPS(s, gnhpsCmd,
		closeChan)
	if err != nil {
		return nil, err
	}
	networkHashesPerSec, ok := networkHashesPerSecIface.(int64)
	if !ok {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCInternal.Code,
			Message: "networkHashesPerSec is not an int64",
		}
	}

	selectedTipHash := s.cfg.DAG.SelectedTipHash()
	selectedBlock, err := s.cfg.DAG.BlockByHash(selectedTipHash)
	if err != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCInternal.Code,
			Message: "could not find block for selected tip",
		}
	}

	result := btcjson.GetMiningInfoResult{
		Blocks:           int64(s.cfg.DAG.BlockCount()),
		CurrentBlockSize: uint64(selectedBlock.MsgBlock().SerializeSize()),
		CurrentBlockTx:   uint64(len(selectedBlock.MsgBlock().Transactions)),
		Difficulty:       getDifficultyRatio(s.cfg.DAG.CurrentBits(), s.cfg.DAGParams),
		Generate:         s.cfg.CPUMiner.IsMining(),
		GenProcLimit:     s.cfg.CPUMiner.NumWorkers(),
		HashesPerSec:     int64(s.cfg.CPUMiner.HashesPerSecond()),
		NetworkHashPS:    networkHashesPerSec,
		PooledTx:         uint64(s.cfg.TxMemPool.Count()),
		TestNet:          config.MainConfig().TestNet,
		DevNet:           config.MainConfig().DevNet,
	}
	return &result, nil
}
