package rpc

import (
	"github.com/kaspanet/kaspad/rpc/model"
	"github.com/kaspanet/kaspad/util/daghash"
)

// handleGetBlockDAGInfo implements the getBlockDagInfo command.
func handleGetBlockDAGInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	// Obtain a snapshot of the current best known DAG state. We'll
	// populate the response to this call primarily from this snapshot.
	params := s.dag.Params
	dag := s.dag

	dagInfo := &model.GetBlockDAGInfoResult{
		DAG:           params.Name,
		Blocks:        dag.BlockCount(),
		Headers:       dag.BlockCount(),
		TipHashes:     daghash.Strings(dag.TipHashes()),
		Difficulty:    getDifficultyRatio(dag.CurrentBits(), params),
		MedianTime:    dag.CalcPastMedianTime().UnixMilliseconds(),
		Pruned:        false,
		Bip9SoftForks: make(map[string]*model.Bip9SoftForkDescription),
	}

	return dagInfo, nil
}
