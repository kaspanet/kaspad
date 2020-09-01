package integration

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/client/grpcclient"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
	"time"
)

const testTimeout = 1 * time.Second

type testRPCClient struct {
	*grpcclient.RPCClient

	rpcAddress   string
	router       *testRPCRouter
	onBlockAdded func(header *appmessage.BlockHeader)
}

func newTestRPCClient(rpcAddress string) (*testRPCClient, error) {
	rpcClient, err := grpcclient.Connect(rpcAddress)
	if err != nil {
		return nil, errors.Wrapf(err, "error connecting to address %s", rpcClient)
	}
	testRouter, err := buildTestRPCRouter()
	if err != nil {
		return nil, errors.Wrapf(err, "error creating the miner router")
	}
	rpcClient.AttachRouter(testRouter.router)

	testClient := &testRPCClient{
		RPCClient:  rpcClient,
		rpcAddress: rpcAddress,
		router:     testRouter,
	}

	err = testClient.registerForBlockAddedNotifications()
	if err != nil {
		return nil, errors.Wrapf(err, "error requesting block-added notifications")
	}

	log.Infof("Connected to server %s", rpcAddress)

	return testClient, nil
}

type testRPCRouter struct {
	router                        *routerpkg.Router
	getBlockTemplateResponseRoute *routerpkg.Route
	submitBlockResponseRoute      *routerpkg.Route
	notifyBlockAddedResponseRoute *routerpkg.Route
	blockAddedNotificationRoute   *routerpkg.Route
}

func buildTestRPCRouter() (*testRPCRouter, error) {
	router := routerpkg.NewRouter()
	getBlockTemplateResponseRoute, err := router.AddIncomingRoute([]appmessage.MessageCommand{appmessage.CmdGetBlockTemplateResponseMessage})
	if err != nil {
		return nil, err
	}
	submitBlockResponseRoute, err := router.AddIncomingRoute([]appmessage.MessageCommand{appmessage.CmdSubmitBlockResponseMessage})
	if err != nil {
		return nil, err
	}
	notifyBlockAddedResponseRoute, err := router.AddIncomingRoute([]appmessage.MessageCommand{appmessage.CmdNotifyBlockAddedResponseMessage})
	if err != nil {
		return nil, err
	}
	blockAddedNotificationRoute, err := router.AddIncomingRoute([]appmessage.MessageCommand{appmessage.CmdBlockAddedNotificationMessage})
	if err != nil {
		return nil, err
	}

	minerRouter := &testRPCRouter{
		router: router,

		getBlockTemplateResponseRoute: getBlockTemplateResponseRoute,
		submitBlockResponseRoute:      submitBlockResponseRoute,
		notifyBlockAddedResponseRoute: notifyBlockAddedResponseRoute,
		blockAddedNotificationRoute:   blockAddedNotificationRoute,
	}

	return minerRouter, nil
}

func (r *testRPCRouter) outgoingRoute() *routerpkg.Route {
	return r.router.OutgoingRoute()
}

func (c *testRPCClient) address() string {
	return c.rpcAddress
}

func (c *testRPCClient) registerForBlockAddedNotifications() error {
	err := c.router.outgoingRoute().Enqueue(appmessage.NewNotifyBlockAddedRequestMessage())
	if err != nil {
		return err
	}
	response, err := c.router.notifyBlockAddedResponseRoute.DequeueWithTimeout(testTimeout)
	if err != nil {
		return err
	}
	notifyBlockAddedResponse := response.(*appmessage.NotifyBlockAddedResponseMessage)
	if notifyBlockAddedResponse.Error != nil {
		return c.convertRPCError(notifyBlockAddedResponse.Error)
	}
	spawn("registerForBlockAddedNotifications-blockAddedNotificationChan", func() {
		for {
			notification, err := c.router.blockAddedNotificationRoute.Dequeue()
			if err != nil {
				panic(err)
			}
			blockAddedNotification := notification.(*appmessage.BlockAddedNotificationMessage)
			c.onBlockAdded(&blockAddedNotification.Block.Header)

		}
	})
	return nil
}

func (c *testRPCClient) submitBlock(block *util.Block) error {
	blockHex := ""
	if block != nil {
		blockBytes, err := block.Bytes()
		if err != nil {
			return err
		}
		blockHex = hex.EncodeToString(blockBytes)
	}
	err := c.router.outgoingRoute().Enqueue(appmessage.NewSubmitBlockRequestMessage(blockHex))
	if err != nil {
		return err
	}
	response, err := c.router.submitBlockResponseRoute.DequeueWithTimeout(testTimeout)
	if err != nil {
		return err
	}
	submitBlockResponse := response.(*appmessage.SubmitBlockResponseMessage)
	if submitBlockResponse.Error != nil {
		return c.convertRPCError(submitBlockResponse.Error)
	}
	return nil
}

func (c *testRPCClient) getBlockTemplate(miningAddress string, longPollID string) (*appmessage.GetBlockTemplateResponseMessage, error) {
	err := c.router.outgoingRoute().Enqueue(appmessage.NewGetBlockTemplateRequestMessage(miningAddress, longPollID))
	if err != nil {
		return nil, err
	}
	response, err := c.router.getBlockTemplateResponseRoute.DequeueWithTimeout(testTimeout)
	if err != nil {
		return nil, err
	}
	getBlockTemplateResponse := response.(*appmessage.GetBlockTemplateResponseMessage)
	if getBlockTemplateResponse.Error != nil {
		return nil, c.convertRPCError(getBlockTemplateResponse.Error)
	}
	return getBlockTemplateResponse, nil
}

func (c *testRPCClient) convertRPCError(rpcError *appmessage.RPCError) error {
	return errors.Errorf("received error response from RPC: %s", rpcError.Message)
}
