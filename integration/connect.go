package integration

import (
	"testing"
	"time"

	"github.com/kaspanet/kaspad/app"
)

func connect(t *testing.T, app1, app2 *app.App, client1, client2 *rpcClient, app1P2PAddress string) {
	err := client2.ConnectNode(app1P2PAddress)
	if err != nil {
		t.Fatalf("Error connecting the nodes")
	}

	onConnectedChan := make(chan struct{})
	abortConnectionChan := make(chan struct{})
	defer func() { close(abortConnectionChan) }()

	spawn("integration.connect-Wait for connection", func() {
		for range time.Tick(10 * time.Millisecond) {
			if isConnected(t, app1, app2, client1, client2) {
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
func isConnected(t *testing.T, app1, app2 *app.App, client1, client2 *rpcClient) bool {
	connectedPeerInfo1, err := client1.GetConnectedPeerInfo()
	if err != nil {
		t.Fatalf("Error getting connected peer info for app1: %+v", err)
	}
	connectedPeerInfo2, err := client2.GetConnectedPeerInfo()
	if err != nil {
		t.Fatalf("Error getting connected peer info for app2: %+v", err)
	}

	var app1Connected, app2Connected bool
	app1ID, app2ID := app1.ID().String(), app2.ID().String()

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

	return app1Connected
}
