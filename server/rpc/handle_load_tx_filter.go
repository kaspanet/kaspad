package rpc

import (
	"github.com/kaspanet/kaspad/rpcmodel"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

// handleLoadTxFilter implements the loadTxFilter command extension for
// websocket connections.
//
// NOTE: This extension is ported from github.com/decred/dcrd
func handleLoadTxFilter(wsc *wsClient, icmd interface{}) (interface{}, error) {
	cmd := icmd.(*rpcmodel.LoadTxFilterCmd)

	outpoints := make([]wire.Outpoint, len(cmd.Outpoints))
	for i := range cmd.Outpoints {
		txID, err := daghash.NewTxIDFromStr(cmd.Outpoints[i].TxID)
		if err != nil {
			return nil, &rpcmodel.RPCError{
				Code:    rpcmodel.ErrRPCInvalidParameter,
				Message: err.Error(),
			}
		}
		outpoints[i] = wire.Outpoint{
			TxID:  *txID,
			Index: cmd.Outpoints[i].Index,
		}
	}

	params := wsc.server.cfg.DAGParams

	wsc.Lock()
	if cmd.Reload || wsc.filterData == nil {
		wsc.filterData = newWSClientFilter(cmd.Addresses, outpoints,
			params)
		wsc.Unlock()
	} else {
		wsc.Unlock()

		func() {
			wsc.filterData.mu.Lock()
			defer wsc.filterData.mu.Unlock()
			for _, a := range cmd.Addresses {
				wsc.filterData.addAddressStr(a, params)
			}
			for i := range outpoints {
				wsc.filterData.addUnspentOutpoint(&outpoints[i])
			}
		}()
	}

	return nil, nil
}
