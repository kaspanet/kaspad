package rpc

import "github.com/kaspanet/kaspad/rpcmodel"

// handleSession implements the session command extension for websocket
// connections.
func handleSession(wsc *wsClient, icmd interface{}) (interface{}, error) {
	return &rpcmodel.SessionResult{SessionID: wsc.sessionID}, nil
}
