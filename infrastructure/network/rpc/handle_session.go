package rpc

import "github.com/kaspanet/kaspad/infrastructure/network/rpc/model"

// handleSession implements the session command extension for websocket
// connections.
func handleSession(wsc *wsClient, icmd interface{}) (interface{}, error) {
	return &model.SessionResult{SessionID: wsc.sessionID}, nil
}
