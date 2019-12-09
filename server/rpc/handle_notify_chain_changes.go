package rpc

import "github.com/kaspanet/kaspad/btcjson"

// handleNotifyChainChanges implements the notifyChainChanges command extension for
// websocket connections.
func handleNotifyChainChanges(wsc *wsClient, icmd interface{}) (interface{}, error) {
	if wsc.server.cfg.AcceptanceIndex == nil {
		return nil, &btcjson.RPCError{
			Code: btcjson.ErrRPCNoAcceptanceIndex,
			Message: "The acceptance index must be " +
				"enabled to receive chain changes " +
				"(specify --acceptanceindex)",
		}
	}

	wsc.server.ntfnMgr.RegisterChainChanges(wsc)
	return nil, nil
}
