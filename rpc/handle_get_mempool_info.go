package rpc

import "github.com/kaspanet/kaspad/rpc/model"

// handleGetMempoolInfo implements the getMempoolInfo command.
func handleGetMempoolInfo(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	mempoolTxns := s.txMempool.TxDescs()

	var numBytes int64
	for _, txD := range mempoolTxns {
		numBytes += int64(txD.Tx.MsgTx().SerializeSize())
	}

	ret := &model.GetMempoolInfoResult{
		Size:  int64(len(mempoolTxns)),
		Bytes: numBytes,
	}

	return ret, nil
}
