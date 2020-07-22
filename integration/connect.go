package integration

import (
	"testing"
	"time"

	"github.com/kaspanet/kaspad/protocol/peer"

	kaspadpkg "github.com/kaspanet/kaspad/kaspad"
)

func connect(t *testing.T, kaspad1, kaspad2 *kaspadpkg.Kaspad, client1, client2 *rpcClient) {
	kaspad1OnConnectedChan := make(chan struct{})
	kaspad1.ProtocolManager.SetPeerAddedCallback(func(peer *peer.Peer) {
		close(kaspad1OnConnectedChan)
	})

	err := client2.ConnectNode(kaspad1P2PAddress)
	if err != nil {
		t.Fatalf("Error connecting the nodes")
	}

	select {
	case <-kaspad1OnConnectedChan:
	case <-time.After(defaultTimeout):
		t.Fatalf("Timed out waiting for the kaspads to connect")
	}

	verifyConnected(t, client1)
	verifyConnected(t, client2)
}
func verifyConnected(t *testing.T, client *rpcClient) {
	connectedPeerInfo, err := client.GetConnectedPeerInfo()
	if err != nil {
		t.Fatalf("Error getting connected peer info from kaspad1")
	}
	if len(connectedPeerInfo) != 1 {
		t.Errorf("Expected 1 connected peer, but got %d", len(connectedPeerInfo))
	}
}
