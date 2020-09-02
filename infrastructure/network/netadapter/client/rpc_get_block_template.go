package client

import "github.com/kaspanet/kaspad/app/appmessage"

func (c *RPCClient) GetBlockTemplate(miningAddress string, longPollID string) (*appmessage.GetBlockTemplateResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetBlockTemplateRequestMessage(miningAddress, longPollID))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetBlockTemplateResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	getBlockTemplateResponse := response.(*appmessage.GetBlockTemplateResponseMessage)
	if getBlockTemplateResponse.Error != nil {
		return nil, c.convertRPCError(getBlockTemplateResponse.Error)
	}
	return getBlockTemplateResponse, nil
}
