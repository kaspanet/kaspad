package rpc

import (
	"time"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"

	"github.com/kaspanet/kaspad/domain/dagconfig"
)

const defaultRPCServer = "localhost"

// RPCConfig are configurations common to all tests that need to connect to json-rpc
type RPCConfig struct {
	RPCServer string `short:"s" long:"rpcserver" description:"RPC server to connect to"`
}

// ValidateRPCConfig makes sure that provided RPCConfig is valid or returns an error otherwise
func ValidateRPCConfig(config *RPCConfig) error {
	if config.RPCServer == "" {
		config.RPCServer = defaultRPCServer
	}
	return nil
}

type RPCClient struct {
	*rpcclient.RPCClient
	OnBlockAdded chan struct{}
}

// ConnectToRPC connects to JSON-RPC server specified in the provided config
func ConnectToRPC(config *RPCConfig, dagParams *dagconfig.Params) (*RPCClient, error) {
	rpcAddress, err := dagParams.NormalizeRPCServerAddress(config.RPCServer)
	if err != nil {
		return nil, err
	}
	rpcClient, err := rpcclient.NewRPCClient(rpcAddress)
	if err != nil {
		return nil, err
	}
	rpcClient.SetTimeout(time.Second * 120)
	rpcClient.SetOnErrorHandler(func(err error) {
		log.Errorf("Error from RPCClient: %+v", err)
	})

	client := &RPCClient{
		RPCClient:    rpcClient,
		OnBlockAdded: make(chan struct{}),
	}

	return client, nil
}

func (c *RPCClient) RegisterForBlockAddedNotifications() error {
	return c.RPCClient.RegisterForBlockAddedNotifications(func(_ *appmessage.BlockAddedNotificationMessage) {
		c.OnBlockAdded <- struct{}{}
	})
}
