package integration

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/client/grpcclient"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
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
		return nil, errors.Wrapf(err, "error creating the test router")
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
	router *routerpkg.Router
	routes map[appmessage.MessageCommand]*routerpkg.Route
}

func buildTestRPCRouter() (*testRPCRouter, error) {
	router := routerpkg.NewRouter()
	routes := make(map[appmessage.MessageCommand]*routerpkg.Route, len(appmessage.RPCMessageCommandToString))
	for messageType := range appmessage.RPCMessageCommandToString {
		route, err := router.AddIncomingRoute([]appmessage.MessageCommand{messageType})
		if err != nil {
			return nil, err
		}
		routes[messageType] = route
	}

	testRPCRouter := &testRPCRouter{
		router: router,
		routes: routes,
	}

	return testRPCRouter, nil
}

func (r *testRPCRouter) outgoingRoute() *routerpkg.Route {
	return r.router.OutgoingRoute()
}

func (c *testRPCClient) address() string {
	return c.rpcAddress
}

func (c *testRPCClient) route(command appmessage.MessageCommand) *routerpkg.Route {
	return c.router.routes[command]
}

func (c *testRPCClient) registerForBlockAddedNotifications() error {
	err := c.router.outgoingRoute().Enqueue(appmessage.NewNotifyBlockAddedRequestMessage())
	if err != nil {
		return err
	}
	response, err := c.route(appmessage.CmdNotifyBlockAddedResponseMessage).DequeueWithTimeout(testTimeout)
	if err != nil {
		return err
	}
	notifyBlockAddedResponse := response.(*appmessage.NotifyBlockAddedResponseMessage)
	if notifyBlockAddedResponse.Error != nil {
		return c.convertRPCError(notifyBlockAddedResponse.Error)
	}
	spawn("registerForBlockAddedNotifications-blockAddedNotificationChan", func() {
		for {
			notification, err := c.route(appmessage.CmdBlockAddedNotificationMessage).Dequeue()
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
	response, err := c.route(appmessage.CmdSubmitBlockResponseMessage).DequeueWithTimeout(testTimeout)
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
	response, err := c.route(appmessage.CmdGetBlockTemplateResponseMessage).DequeueWithTimeout(testTimeout)
	if err != nil {
		return nil, err
	}
	getBlockTemplateResponse := response.(*appmessage.GetBlockTemplateResponseMessage)
	if getBlockTemplateResponse.Error != nil {
		return nil, c.convertRPCError(getBlockTemplateResponse.Error)
	}
	return getBlockTemplateResponse, nil
}

func (c *testRPCClient) getPeerAddresses() (*appmessage.GetPeerAddressesResponseMessage, error) {
	err := c.router.outgoingRoute().Enqueue(appmessage.NewGetPeerAddressesRequestMessage())
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetPeerAddressesResponseMessage).DequeueWithTimeout(testTimeout)
	if err != nil {
		return nil, err
	}
	getPeerAddressesResponse := response.(*appmessage.GetPeerAddressesResponseMessage)
	if getPeerAddressesResponse.Error != nil {
		return nil, c.convertRPCError(getPeerAddressesResponse.Error)
	}
	return getPeerAddressesResponse, nil
}

func (c *testRPCClient) getSelectedTipHash() (*appmessage.GetSelectedTipHashResponseMessage, error) {
	err := c.router.outgoingRoute().Enqueue(appmessage.NewGetSelectedTipHashRequestMessage())
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetSelectedTipHashResponseMessage).DequeueWithTimeout(testTimeout)
	if err != nil {
		return nil, err
	}
	getSelectedTipHashResponse := response.(*appmessage.GetSelectedTipHashResponseMessage)
	if getSelectedTipHashResponse.Error != nil {
		return nil, c.convertRPCError(getSelectedTipHashResponse.Error)
	}
	return getSelectedTipHashResponse, nil
}

func (c *testRPCClient) sendRawTransaction(tx *appmessage.MsgTx, allowHighFees bool) (*daghash.TxID, error) {
	return nil, nil
}

func (c *testRPCClient) getMempoolEntry(txID string) (*appmessage.GetMempoolEntryResponseMessage, error) {
	err := c.router.outgoingRoute().Enqueue(appmessage.NewGetMempoolEntryRequestMessage(txID))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetMempoolEntryResponseMessage).DequeueWithTimeout(testTimeout)
	if err != nil {
		return nil, err
	}
	getMempoolEntryResponse := response.(*appmessage.GetMempoolEntryResponseMessage)
	if getMempoolEntryResponse.Error != nil {
		return nil, c.convertRPCError(getMempoolEntryResponse.Error)
	}
	return getMempoolEntryResponse, nil
}

func (c *testRPCClient) connectNode(host string) error {
	return nil
}

func (c *testRPCClient) getConnectedPeerInfo() (*appmessage.GetConnectedPeerInfoResponseMessage, error) {
	err := c.router.outgoingRoute().Enqueue(appmessage.NewGetConnectedPeerInfoRequestMessage())
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetConnectedPeerInfoResponseMessage).DequeueWithTimeout(testTimeout)
	if err != nil {
		return nil, err
	}
	getMempoolEntryResponse := response.(*appmessage.GetConnectedPeerInfoResponseMessage)
	if getMempoolEntryResponse.Error != nil {
		return nil, c.convertRPCError(getMempoolEntryResponse.Error)
	}
	return getMempoolEntryResponse, nil
}

func (c *testRPCClient) convertRPCError(rpcError *appmessage.RPCError) error {
	return errors.Errorf("received error response from RPC: %s", rpcError.Message)
}
