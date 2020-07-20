package rpc

import "github.com/kaspanet/kaspad/rpc/model"

// handleGetRawMempool implements the getRawMempool command.
func handleGetRawMempool(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*model.GetRawMempoolCmd)
	mp := s.txMempool

	if c.Verbose != nil && *c.Verbose {
		return mp.RawMempoolVerbose(), nil
	}

	// The response is simply an array of the transaction hashes if the
	// verbose flag is not set.
	descs := mp.TxDescs()
	hashStrings := make([]string, len(descs))
	for i := range hashStrings {
		hashStrings[i] = descs[i].Tx.ID().String()
	}

	return hashStrings, nil
}
