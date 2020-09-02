package integration

import (
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"time"
)

const testTimeout = 1 * time.Second

type testRPCClient struct {
	*rpcclient.RPCClient
}

func newTestRPCClient(rpcAddress string) (*testRPCClient, error) {
	rpcClient, err := rpcclient.NewRPCClient(rpcAddress)
	if err != nil {
		return nil, err
	}
	rpcClient.SetTimeout(testTimeout)

	return &testRPCClient{
		RPCClient: rpcClient,
	}, nil
}
