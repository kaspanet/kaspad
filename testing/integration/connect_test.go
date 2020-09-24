package integration

import (
	"testing"
	"time"
)

func connect(t *testing.T, incoming, outgoing *appHarness) {
	err := outgoing.rpcClient.AddPeer(incoming.p2pAddress, false)
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
func isConnected(t *testing.T, appHarness1, appHarness2 *appHarness) bool {
	connectedPeerInfo1, err := appHarness1.rpcClient.GetConnectedPeerInfo()
	if err != nil {
		t.Fatalf("Error getting connected peer info for app1: %+v", err)
	}
	connectedPeerInfo2, err := appHarness2.rpcClient.GetConnectedPeerInfo()
	if err != nil {
		t.Fatalf("Error getting connected peer info for app2: %+v", err)
	}

	var incomingConnected, outgoingConnected bool
	app1ID, app2ID := appHarness1.app.P2PNodeID().String(), appHarness2.app.P2PNodeID().String()

	for _, connectedPeer := range connectedPeerInfo1.Infos {
		if connectedPeer.ID == app2ID {
			incomingConnected = true
			break
		}
	}

	for _, connectedPeer := range connectedPeerInfo2.Infos {
		if connectedPeer.ID == app1ID {
			outgoingConnected = true
			break
		}
	}

	return incomingConnected && outgoingConnected
}
