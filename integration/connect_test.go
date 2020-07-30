package integration

import (
	"testing"
	"time"
)

func connect(t *testing.T, incoming, outgoing *appHarness) {
	err := outgoing.rpcClient.ConnectNode(incoming.p2pAddress)
	if err != nil {
		t.Fatalf("Error connecting the nodes")
	}

	onConnectedChan := make(chan struct{})
	abortConnectionChan := make(chan struct{})
	defer close(abortConnectionChan)

	spawn("integration.connect-Wait for connection", func() {
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			if isConnected(t, incoming, outgoing) {
				close(onConnectedChan)
				return
			}

			select {
			case <-abortConnectionChan:
				return
			default:
			}
		}
	})

	select {
	case <-onConnectedChan:
	case <-time.After(defaultTimeout):
		t.Fatalf("Timed out waiting for the apps to connect")
	}
}
func isConnected(t *testing.T, incoming, outgoing *appHarness) bool {
	connectedPeerInfo1, err := incoming.rpcClient.GetConnectedPeerInfo()
	if err != nil {
		t.Fatalf("Error getting connected peer info for app1: %+v", err)
	}
	connectedPeerInfo2, err := outgoing.rpcClient.GetConnectedPeerInfo()
	if err != nil {
		t.Fatalf("Error getting connected peer info for app2: %+v", err)
	}

	var incomingConnected, outgoingConnected bool
	app1ID, app2ID := incoming.app.P2PNodeID().String(), outgoing.app.P2PNodeID().String()

	for _, connectedPeer := range connectedPeerInfo1 {
		if connectedPeer.ID == app2ID {
			incomingConnected = true
			break
		}
	}

	for _, connectedPeer := range connectedPeerInfo2 {
		if connectedPeer.ID == app1ID {
			outgoingConnected = true
			break
		}
	}

	return incomingConnected && outgoingConnected
}
