package integration

import (
	"testing"
	"time"

	rpcclient "github.com/kaspanet/kaspad/rpc/client"
)

func TestIntegration(t *testing.T) {
	_, _, client1, client2, teardown := setup(t)

	// Wait for half a second to let things start and connect
	<-time.After(time.Second / 2)

	defer teardown()
	verifyConnected(t, client1, kaspad2P2PAddress)
	verifyConnected(t, client2, kaspad1P2PAddress)
}

func verifyConnected(t *testing.T, client *rpcclient.Client, expectedAddress string) {
	connectedPeerInfo, err := client.GetConnectedPeerInfo()
	if err != nil {
		t.Fatalf("Error getting connected peer info from kaspad1")
	}
	if len(connectedPeerInfo) != 1 {
		t.Errorf("Expected 1 connected peer, but got %d", len(connectedPeerInfo))
	}
}
