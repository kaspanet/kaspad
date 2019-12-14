package rpc

import "github.com/kaspanet/kaspad/rpcmodel"

// handleNotifyChainChanges implements the notifyChainChanges command extension for
// websocket connections.
func handleNotifyChainChanges(wsc *wsClient, icmd interface{}) (interface{}, error) {
	if wsc.server.cfg.AcceptanceIndex == nil {
		return nil, &rpcmodel.RPCError{
			Code: rpcmodel.ErrRPCNoAcceptanceIndex,
			Message: "The acceptance index must be " +
				"enabled to receive chain changes " +
				"(specify --acceptanceindex)",
		}
	}

	wsc.server.ntfnMgr.RegisterChainChanges(wsc)
	return nil, nil
}
