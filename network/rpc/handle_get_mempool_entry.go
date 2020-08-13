package rpc

import (
	"github.com/kaspanet/kaspad/network/rpc/model"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/pkg/errors"
)

func handleGetMempoolEntry(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*model.GetMempoolEntryCmd)
	txID, err := daghash.NewTxIDFromStr(c.TxID)
	if err != nil {
		return nil, err
	}

	txDesc, ok := s.txMempool.FetchTxDesc(txID)
	if !ok {
		return nil, errors.Errorf("transaction is not in the pool")
	}

	tx := txDesc.Tx
	rawTx, err := createTxRawResult(s.dag.Params, tx.MsgTx(), tx.ID().String(),
		nil, "", nil, true)
	if err != nil {
		return nil, err
	}

	return &model.GetMempoolEntryResult{
		Fee:   txDesc.Fee,
		Time:  txDesc.Added.UnixMilliseconds(),
		RawTx: *rawTx,
	}, nil
}
