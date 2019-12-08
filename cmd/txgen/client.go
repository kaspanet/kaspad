package main

import (
	"github.com/kaspanet/kaspad/rpcclient"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
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
		return nil, errors.Errorf("Error connecting to address %s: %s", connCfg.Host, err)
	}

	if err = client.NotifyBlocks(); err != nil {
		return nil, errors.Errorf("Error while registering client %s for block notifications: %s", client.Host(), err)
	}
	return client, nil
}
