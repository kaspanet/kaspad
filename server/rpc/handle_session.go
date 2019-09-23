package rpc

import "github.com/daglabs/btcd/btcjson"

// handleSession implements the session command extension for websocket
// connections.
func handleSession(wsc *wsClient, icmd interface{}) (interface{}, error) {
	return &btcjson.SessionResult{SessionID: wsc.sessionID}, nil
}
