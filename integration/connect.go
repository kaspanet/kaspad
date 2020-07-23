package integration

import (
	"testing"
	"time"

	"github.com/kaspanet/kaspad/app"
	"github.com/kaspanet/kaspad/protocol/peer"
)

func connect(t *testing.T, app1, app2 *app.App, client1, client2 *rpcClient) {
	app1OnConnectedChan := make(chan struct{})
	app1.ProtocolManager.SetPeerAddedCallback(func(peer *peer.Peer) {
		close(app1OnConnectedChan)
	})

	err := client2.ConnectNode(p2pAddress1)
	if err != nil {
		t.Fatalf("Error connecting the nodes")
	}

	select {
	case <-app1OnConnectedChan:
	case <-time.After(defaultTimeout):
		t.Fatalf("Timed out waiting for the apps to connect")
	}

	verifyConnected(t, client1)
	verifyConnected(t, client2)
}
func verifyConnected(t *testing.T, client *rpcClient) {
	connectedPeerInfo, err := client.GetConnectedPeerInfo()
	if err != nil {
		t.Fatalf("Error getting connected peer info: %+v", err)
	}
	if len(connectedPeerInfo) != 1 {
		t.Errorf("Expected 1 connected peer, but got %d", len(connectedPeerInfo))
	}
}
