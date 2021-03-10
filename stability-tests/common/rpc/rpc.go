package rpc

import (
	"time"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"

	"github.com/kaspanet/kaspad/domain/dagconfig"
)

const defaultRPCServer = "localhost"

// Config are configurations common to all tests that need to connect to json-rpc
type Config struct {
	RPCServer string `short:"s" long:"rpcserver" description:"RPC server to connect to"`
}

// ValidateRPCConfig makes sure that provided Config is valid or returns an error otherwise
func ValidateRPCConfig(config *Config) error {
	if config.RPCServer == "" {
		config.RPCServer = defaultRPCServer
	}
	return nil
}

// Client wraps rpcclient.RPCClient with extra functionality needed for stability-tests
type Client struct {
	*rpcclient.RPCClient
	OnBlockAdded chan struct{}
}

// ConnectToRPC connects to JSON-RPC server specified in the provided config
func ConnectToRPC(config *Config, dagParams *dagconfig.Params) (*Client, error) {
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
		log.Errorf("Error from Client: %+v", err)
	})

	client := &Client{
		RPCClient:    rpcClient,
		OnBlockAdded: make(chan struct{}),
	}

	return client, nil
}

// RegisterForBlockAddedNotifications registers for block added notifications pushed by the node
func (c *Client) RegisterForBlockAddedNotifications() error {
	return c.RPCClient.RegisterForBlockAddedNotifications(func(_ *appmessage.BlockAddedNotificationMessage) {
		c.OnBlockAdded <- struct{}{}
	})
}
