package rpc

import "github.com/daglabs/btcd/btcjson"

// handleGetBestBlock implements the getBestBlock command.
func handleGetBestBlock(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	// All other "get block" commands give either the height, the
	// hash, or both but require the block SHA.  This gets both for
	// the best block.
	result := &btcjson.GetBestBlockResult{
		Hash:   s.cfg.DAG.SelectedTipHash().String(),
		Height: s.cfg.DAG.ChainHeight(), //TODO: (Ori) This is probably wrong. Done only for compilation
	}
	return result, nil
}
