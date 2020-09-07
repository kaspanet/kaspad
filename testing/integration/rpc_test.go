package integration

import (
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"time"
)

const rpcTimeout = 1 * time.Second

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
