package main

import (
	"github.com/kaspanet/kaspad/rpcclient"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

type simulatorClient struct {
	*rpcclient.Client
	onBlockAdded       chan struct{}
	notifyForNewBlocks bool
}

func newSimulatorClient(address string, connCfg *rpcclient.ConnConfig) (*simulatorClient, error) {
	client := &simulatorClient{
		onBlockAdded: make(chan struct{}, 1),
	}
	notificationHandlers := &rpcclient.NotificationHandlers{
		OnFilteredBlockAdded: func(height uint64, header *wire.BlockHeader,
			txs []*util.Tx) {
			if client.notifyForNewBlocks {
				client.onBlockAdded <- struct{}{}
			}
		},
	}
	var err error
	client.Client, err = rpcclient.New(connCfg, notificationHandlers)
	if err != nil {
		return nil, errors.Errorf("Error connecting to address %s: %s", address, err)
	}

	if err = client.NotifyBlocks(); err != nil {
		return nil, errors.Errorf("Error while registering client %s for block notifications: %s", client.Host(), err)
	}
	return client, nil
}
