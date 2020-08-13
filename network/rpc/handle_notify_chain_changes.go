package rpc

import "github.com/kaspanet/kaspad/network/rpc/model"

// handleNotifyChainChanges implements the notifyChainChanges command extension for
// websocket connections.
func handleNotifyChainChanges(wsc *wsClient, icmd interface{}) (interface{}, error) {
	if wsc.server.acceptanceIndex == nil {
		return nil, &model.RPCError{
			Code: model.ErrRPCNoAcceptanceIndex,
			Message: "The acceptance index must be " +
				"enabled to receive chain changes " +
				"(specify --acceptanceindex)",
		}
	}

	wsc.server.ntfnMgr.RegisterChainChanges(wsc)
	return nil, nil
}
