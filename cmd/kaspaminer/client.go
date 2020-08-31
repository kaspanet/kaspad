package main

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server/grpcserver/grpcclient"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
)

type minerClient struct {
	*grpcclient.RPCClient

	rpcAddress   string
	onBlockAdded chan struct{}
}

func newMinerClient(cfg *configFlags) (*minerClient, error) {
	rpcAddress, err := cfg.NetParams().NormalizeRPCServerAddress(cfg.RPCServer)
	if err != nil {
		return nil, err
	}
	rpcClient, err := connectToServer(rpcAddress)
	if err != nil {
		return nil, err
	}

	minerClient := &minerClient{
		RPCClient: rpcClient,

		rpcAddress:   rpcAddress,
		onBlockAdded: make(chan struct{}, 1),
	}
	//notificationHandlers := &client.NotificationHandlers{
	//	OnFilteredBlockAdded: func(_ uint64, header *appmessage.BlockHeader,
	//		txs []*util.Tx) {
	//		minerClient.onBlockAdded <- struct{}{}
	//	},
	//}

	return minerClient, nil
}

func connectToServer(rpcAddress string) (*grpcclient.RPCClient, error) {
	rpcClient, err := grpcclient.Connect(rpcAddress)
	if err != nil {
		return nil, errors.Wrapf(err, "error connecting to address %s", rpcClient)
	}
	_, err = rpcClient.PostAppMessage(appmessage.NewNotifyBlockAddedRequestMessage())
	if err != nil {
		return nil, errors.Wrapf(err, "error requesting block-added notifications")
	}

	log.Infof("Connected to server %s", rpcAddress)

	return rpcClient, nil
}

func (c *minerClient) Address() string {
	return c.rpcAddress
}

func (c *minerClient) SubmitBlock(block *util.Block) error {
	blockHex := ""
	if block != nil {
		blockBytes, err := block.Bytes()
		if err != nil {
			return err
		}
		blockHex = hex.EncodeToString(blockBytes)
	}
	_, err := c.PostAppMessage(appmessage.NewSubmitBlockRequestMessage(blockHex))
	return err
}

func (c *minerClient) GetBlockTemplate(miningAddress string, longPollID string) (*appmessage.GetBlockTemplateResponseMessage, error) {
	response, err := c.PostAppMessage(appmessage.NewGetBlockTemplateRequestMessage(miningAddress, longPollID))
	if err != nil {
		return nil, err
	}
	return response.(*appmessage.GetBlockTemplateResponseMessage), nil
}
