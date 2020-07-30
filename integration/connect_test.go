
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
		for range time.Tick(10 * time.Millisecond) {
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

	var app1Connected, app2Connected bool
	app1ID, app2ID := appHarness1.app.P2PNodeID().String(), appHarness2.app.P2PNodeID().String()

	for _, connectedPeer := range connectedPeerInfo1 {
		if connectedPeer.ID == app2ID {
			app1Connected = true
			break
		}
	}

	for _, connectedPeer := range connectedPeerInfo2 {
		if connectedPeer.ID == app1ID {
			app2Connected = true
			break
		}
	}

	if (app1Connected && !app2Connected) || (!app1Connected && app2Connected) {
		t.Fatalf("app1Connected is %t while app2Connected is %t", app1Connected, app2Connected)
	}

	return app1Connected && app2Connected
}
