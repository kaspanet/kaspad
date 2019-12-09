package rpc

import "github.com/kaspanet/kaspad/kaspajson"

// handleSession implements the session command extension for websocket
// connections.
func handleSession(wsc *wsClient, icmd interface{}) (interface{}, error) {
	return &kaspajson.SessionResult{SessionID: wsc.sessionID}, nil
}
