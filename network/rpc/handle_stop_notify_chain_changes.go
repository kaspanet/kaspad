package rpc

// handleStopNotifyChainChanges implements the stopNotifyChainChanges command extension for
// websocket connections.
func handleStopNotifyChainChanges(wsc *wsClient, icmd interface{}) (interface{}, error) {
	wsc.server.ntfnMgr.UnregisterChainChanges(wsc)
	return nil, nil
}
