package integration

import (
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server/grpcserver"
	"testing"
	"time"

	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
)

const rpcTimeout = 10 * time.Second

type testRPCClient struct {
	*rpcclient.RPCClient
}

func newTestRPCClient(rpcAddress string) (*testRPCClient, error) {
	rpcClient, err := rpcclient.NewRPCClient(rpcAddress)
	if err != nil {
		return nil, err
	}
	rpcClient.SetTimeout(rpcTimeout)

	return &testRPCClient{
		RPCClient: rpcClient,
	}, nil
}

func TestRPCMaxInboundConnections(t *testing.T) {
	harness, teardown := setupHarness(t, &harnessParams{
		p2pAddress:              p2pAddress1,
		rpcAddress:              rpcAddress1,
		miningAddress:           miningAddress1,
		miningAddressPrivateKey: miningAddress1PrivateKey,
	})
	defer teardown()

	// Close the default RPC client so that it won't interfere with the test
	err := harness.rpcClient.Close()
	if err != nil {
		t.Fatalf("Failed to close the default harness RPCClient: %s", err)
	}

	// Connect `RPCMaxInboundConnections` clients. We expect this to succeed immediately
	rpcClients := []*testRPCClient{}
	doneChan := make(chan error)
	go func() {
		for i := 0; i < grpcserver.RPCMaxInboundConnections; i++ {
			rpcClient, err := newTestRPCClient(harness.rpcAddress)
			if err != nil {
				doneChan <- err
			}
			rpcClients = append(rpcClients, rpcClient)
		}
		doneChan <- nil
	}()
	select {
	case err = <-doneChan:
		if err != nil {
			t.Fatalf("newTestRPCClient: %s", err)
		}
	case <-time.After(time.Second * 5):
		t.Fatalf("Timeout for connecting %d RPC connections elapsed", grpcserver.RPCMaxInboundConnections)
	}

	// Try to connect another client. We expect this to fail
	// We set a timeout to account for reconnection mechanisms
	go func() {
		rpcClient, err := newTestRPCClient(harness.rpcAddress)
		if err != nil {
			doneChan <- err
		}
		rpcClients = append(rpcClients, rpcClient)
		doneChan <- nil
	}()
	select {
	case err = <-doneChan:
		if err == nil {
			t.Fatalf("newTestRPCClient unexpectedly succeeded")
		}
	case <-time.After(time.Second * 15):
	}
}
