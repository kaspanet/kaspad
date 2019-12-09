package rpc

import "github.com/kaspanet/kaspad/jsonrpc"

// handleSession implements the session command extension for websocket
// connections.
func handleSession(wsc *wsClient, icmd interface{}) (interface{}, error) {
	return &jsonrpc.SessionResult{SessionID: wsc.sessionID}, nil
}
