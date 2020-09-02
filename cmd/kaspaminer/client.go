package main

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/client"
	"github.com/pkg/errors"
	"time"
)

const minerTimeout = 10 * time.Second

type minerClient struct {
	*client.RPCClient

	blockAddedNotificationChan chan struct{}
}

func newMinerClient(cfg *configFlags) (*minerClient, error) {
	rpcAddress, err := cfg.NetParams().NormalizeRPCServerAddress(cfg.RPCServer)
	if err != nil {
		return nil, err
	}
	rpcClient, err := client.NewRPCClient(rpcAddress)
	if err != nil {
		return nil, err
	}
	rpcClient.SetTimeout(minerTimeout)

	minerClient := &minerClient{
		RPCClient:                  rpcClient,
		blockAddedNotificationChan: make(chan struct{}),
	}

	err = rpcClient.RegisterForBlockAddedNotifications(func(_ *appmessage.BlockHeader) {
		select {
		case minerClient.blockAddedNotificationChan <- struct{}{}:
		default:
		}
	})
	if err != nil {
		return nil, errors.Wrapf(err, "error requesting block-added notifications")
	}

	return minerClient, nil
}
