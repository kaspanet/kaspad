package rpc

import (
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/rpcmodel"
)

// handleGetMiningInfo implements the getMiningInfo command. We only return the
// fields that are not related to wallet functionality.
func handleGetMiningInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	if config.ActiveConfig().SubnetworkID != nil {
		return nil, &rpcmodel.RPCError{
			Code:    rpcmodel.ErrRPCInvalidRequest.Code,
			Message: "`getMiningInfo` is not supported on partial nodes.",
		}
	}

	selectedTipHash := s.cfg.DAG.SelectedTipHash()
	selectedBlock, err := s.cfg.DAG.BlockByHash(selectedTipHash)
	if err != nil {
		return nil, &rpcmodel.RPCError{
			Code:    rpcmodel.ErrRPCInternal.Code,
			Message: "could not find block for selected tip",
		}
	}

	result := rpcmodel.GetMiningInfoResult{
		Blocks:           int64(s.cfg.DAG.BlockCount()),
		CurrentBlockSize: uint64(selectedBlock.MsgBlock().SerializeSize()),
		CurrentBlockTx:   uint64(len(selectedBlock.MsgBlock().Transactions)),
		Difficulty:       getDifficultyRatio(s.cfg.DAG.CurrentBits(), s.cfg.DAGParams),
		Generate:         s.cfg.CPUMiner.IsMining(),
		GenProcLimit:     s.cfg.CPUMiner.NumWorkers(),
		HashesPerSec:     int64(s.cfg.CPUMiner.HashesPerSecond()),
		PooledTx:         uint64(s.cfg.TxMemPool.Count()),
		Testnet:          config.ActiveConfig().Testnet,
		Devnet:           config.ActiveConfig().Devnet,
	}
	return &result, nil
}
