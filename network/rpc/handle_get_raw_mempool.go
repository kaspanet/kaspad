package rpc

import (
	"github.com/kaspanet/kaspad/network/rpc/model"
	"github.com/kaspanet/kaspad/util"
)

// handleGetRawMempool implements the getRawMempool command.
func handleGetRawMempool(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*model.GetRawMempoolCmd)
	mp := s.txMempool

	if c.Verbose != nil && *c.Verbose {
		return rawMempoolVerbose(s), nil
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

// rawMempoolVerbose returns all of the entries in the mempool as a fully
// populated jsonrpc result.
func rawMempoolVerbose(s *Server) map[string]*model.GetRawMempoolVerboseResult {
	descs := s.txMempool.TxDescs()
	result := make(map[string]*model.GetRawMempoolVerboseResult, len(descs))

	for _, desc := range descs {
		// Calculate the current priority based on the inputs to
		// the transaction. Use zero if one or more of the
		// input transactions can't be found for some reason.
		tx := desc.Tx

		mpd := &model.GetRawMempoolVerboseResult{
			Size:    int32(tx.MsgTx().SerializeSize()),
			Fee:     util.Amount(desc.Fee).ToKAS(),
			Time:    desc.Added.UnixMilliseconds(),
			Depends: make([]string, 0),
		}
		for _, txIn := range tx.MsgTx().TxIn {
			txID := &txIn.PreviousOutpoint.TxID
			if s.txMempool.HaveTransaction(txID) {
				mpd.Depends = append(mpd.Depends,
					txID.String())
			}
		}

		result[tx.ID().String()] = mpd
	}

	return result
}
