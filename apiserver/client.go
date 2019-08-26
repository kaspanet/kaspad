package main

import (
	"fmt"
	"github.com/daglabs/btcd/btcjson"
	"github.com/daglabs/btcd/util/daghash"

	"github.com/daglabs/btcd/rpcclient"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/wire"
)

type apiServerClient struct {
	*rpcclient.Client
	onBlockAdded   chan *blockAddedMsg
	onChainChanged chan *chainChangedMsg
}

type blockAddedMsg struct {
	chainHeight uint64
	header      *wire.BlockHeader
}

type chainChangedMsg struct {
	removedChainBlockHashes []*daghash.Hash
	addedChainBlocks        []*btcjson.ChainBlock
}

func newApiServerClient(connCfg *rpcclient.ConnConfig) (*apiServerClient, error) {
	client := &apiServerClient{
		onBlockAdded:   make(chan *blockAddedMsg),
		onChainChanged: make(chan *chainChangedMsg),
	}
	notificationHandlers := &rpcclient.NotificationHandlers{
		OnFilteredBlockAdded: func(height uint64, header *wire.BlockHeader,
			txs []*util.Tx) {
			client.onBlockAdded <- &blockAddedMsg{
				chainHeight: height,
				header:      header,
			}
		},
		OnChainChanged: func(removedChainBlockHashes []*daghash.Hash,
			addedChainBlocks []*btcjson.ChainBlock) {
			client.onChainChanged <- &chainChangedMsg{
				removedChainBlockHashes: removedChainBlockHashes,
				addedChainBlocks:        addedChainBlocks,
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
	if err = client.NotifyChainChanges(); err != nil {
		return nil, fmt.Errorf("Error while registering client %s for chain changes notifications: %s", client.Host(), err)
	}
	return client, nil
}
