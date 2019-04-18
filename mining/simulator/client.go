package main

import (
	"fmt"

	"github.com/daglabs/btcd/rpcclient"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/wire"
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
	ntfnHandlers := &rpcclient.NotificationHandlers{
		OnFilteredBlockAdded: func(height int32, header *wire.BlockHeader,
			txs []*util.Tx) {
			if client.notifyForNewBlocks {
				client.onBlockAdded <- struct{}{}
			}
		},
	}
	var err error
	client.Client, err = rpcclient.New(connCfg, ntfnHandlers)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to address %s: %s", address, err)
	}

	if err = client.NotifyBlocks(); err != nil {
		return nil, fmt.Errorf("Error while registering client %s for block notifications: %s", client.Host(), err)
	}
	return client, nil
}
