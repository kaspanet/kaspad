package main

import (
	"fmt"

	"github.com/daglabs/btcd/rpcclient"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/wire"
)

type txgenClient struct {
	*rpcclient.Client
	onBlockAdded chan *blockAddedMsg
}

type blockAddedMsg struct {
	chainHeight uint64
	header      *wire.BlockHeader
	txs         []*util.Tx
}

func newTxgenClient(connCfg *rpcclient.ConnConfig) (*txgenClient, error) {
	client := &txgenClient{
		onBlockAdded: make(chan *blockAddedMsg),
	}
	notificationHandlers := &rpcclient.NotificationHandlers{
		OnFilteredBlockAdded: func(height uint64, header *wire.BlockHeader,
			txs []*util.Tx) {
			client.onBlockAdded <- &blockAddedMsg{
				chainHeight: height,
				header:      header,
				txs:         txs,
			}
		},
	}
	var err error
	client.Client, err = rpcclient.New(connCfg, notificationHandlers)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to address %s: %s", connCfg.Host, err)
	}

	if err = client.NotifyBlocks(); err != nil {
		return nil, fmt.Errorf("Error while registering client %s for block notifications: %s", client.Host(), err)
	}
	return client, nil
}
