package rpc

// handleNotifyBlocks implements the notifyBlocks command extension for
// websocket connections.
func handleNotifyBlocks(wsc *wsClient, icmd interface{}) (interface{}, error) {
	wsc.server.ntfnMgr.RegisterBlockUpdates(wsc)
	return nil, nil
}
