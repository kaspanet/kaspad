package main

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/client/grpcclient"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
	"time"
)

type minerClient struct {
	*grpcclient.RPCClient

	rpcAddress                 string
	router                     *minerRouter
	blockAddedNotificationChan chan struct{}
}

func newMinerClient(cfg *configFlags) (*minerClient, error) {
	rpcAddress, err := cfg.NetParams().NormalizeRPCServerAddress(cfg.RPCServer)
	if err != nil {
		return nil, err
	}
	rpcClient, err := grpcclient.Connect(rpcAddress)
	if err != nil {
		return nil, errors.Wrapf(err, "error connecting to address %s", rpcClient)
	}
	minerRouter, err := newRouter(rpcClient)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating the miner router")
	}

	minerClient := &minerClient{
		RPCClient:  rpcClient,
		rpcAddress: rpcAddress,
		router:     minerRouter,
	}

	err = minerClient.registerForBlockAddedNotifications()
	if err != nil {
		return nil, errors.Wrapf(err, "error requesting block-added notifications")
	}

	log.Infof("Connected to server %s", rpcAddress)

	return minerClient, nil
}

func (c *minerClient) address() string {
	return c.rpcAddress
}

func (c *minerClient) registerForBlockAddedNotifications() error {
	err := c.router.outgoingRoute().Enqueue(appmessage.NewNotifyBlockAddedRequestMessage())
	if err != nil {
		return err
	}
	spawn("registerForBlockAddedNotifications-blockAddedNotificationChan", func() {
		_, err := c.router.blockAddedNotificationRoute.Dequeue()
		if err != nil {
			panic(err)
		}
		c.blockAddedNotificationChan <- struct{}{}
	})
	return nil
}

func (c *minerClient) submitBlock(block *util.Block) error {
	blockHex := ""
	if block != nil {
		blockBytes, err := block.Bytes()
		if err != nil {
			return err
		}
		blockHex = hex.EncodeToString(blockBytes)
	}
	return c.router.outgoingRoute().Enqueue(appmessage.NewSubmitBlockRequestMessage(blockHex))
}

func (c *minerClient) getBlockTemplate(miningAddress string, longPollID string) (*appmessage.GetBlockTemplateResponseMessage, error) {
	err := c.router.outgoingRoute().Enqueue(appmessage.NewGetBlockTemplateRequestMessage(miningAddress, longPollID))
	if err != nil {
		return nil, err
	}
	response, err := c.router.getBlockTemplateResponseRoute.DequeueWithTimeout(10 * time.Second)
	if err != nil {
		return nil, err
	}
	return response.(*appmessage.GetBlockTemplateResponseMessage), nil
}
