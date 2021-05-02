package integration

import (
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
