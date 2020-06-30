package rpc

import (
	"github.com/kaspanet/kaspad/rpcmodel"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/mstime"
)

func handleGetMempoolEntry(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*rpcmodel.GetMempoolEntryCmd)
	txID, err := daghash.NewTxIDFromStr(c.TxID)
	if err != nil {
		return nil, err
	}

	txDesc, err := s.cfg.TxMemPool.FetchTxDesc(txID)
	if err != nil {
		return nil, err
	}

	tx := txDesc.Tx
	rawTx, err := createTxRawResult(s.cfg.DAGParams, tx.MsgTx(), tx.ID().String(),
		nil, "", nil, true)
	if err != nil {
		return nil, err
	}

	return &rpcmodel.GetMempoolEntryResult{
		Fee:   txDesc.Fee,
		Time:  mstime.TimeToUnixMilli(txDesc.Added),
		RawTx: *rawTx,
	}, nil
}
