package integration

import (
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/client"
	"time"
)

const testTimeout = 1 * time.Second

type testRPCClient struct {
	*client.RPCClient
}

func newTestRPCClient(rpcAddress string) (*testRPCClient, error) {
	rpcClient, err := client.NewRPCClient(rpcAddress)
	if err != nil {
		return nil, err
	}
	rpcClient.SetTimeout(testTimeout)

	return &testRPCClient{
		RPCClient: rpcClient,
	}, nil
}
