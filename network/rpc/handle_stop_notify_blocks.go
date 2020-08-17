package rpc

// handleStopNotifyBlocks implements the stopNotifyBlocks command extension for
// websocket connections.
func handleStopNotifyBlocks(wsc *wsClient, icmd interface{}) (interface{}, error) {
	wsc.server.ntfnMgr.UnregisterBlockUpdates(wsc)
	return nil, nil
}
